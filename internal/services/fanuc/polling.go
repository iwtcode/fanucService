package fanuc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/domain/entities"
)

func (s *Service) StartPolling(ctx context.Context, machineID string, intervalMs int) error {
	// 1. Check if already polling
	if _, exists := s.pollingCancel.Load(machineID); exists {
		return fmt.Errorf("polling already active for machine %s", machineID)
	}

	machine, err := s.CheckConnection(ctx, machineID)
	if err != nil {
		return fmt.Errorf("cannot start polling, machine unreachable: %w", err)
	}

	// 2. Update Mode & Interval in DB
	s.updateInterval(machine, intervalMs)
	s.updateMode(machine, entities.ModePolling)

	// 3. Start polling routine
	s.startPollingInternal(machineID, intervalMs)

	return nil
}

func (s *Service) StopPolling(ctx context.Context, machineID string) error {
	val, ok := s.pollingCancel.Load(machineID)
	if !ok {
		if machine, err := s.repo.GetByID(machineID); err == nil {
			s.updateMode(machine, entities.ModeStatic)
		}
		return fmt.Errorf("polling not active for machine %s", machineID)
	}

	cancel := val.(context.CancelFunc)
	cancel()
	s.pollingCancel.Delete(machineID)

	machine, err := s.repo.GetByID(machineID)
	if err == nil {
		s.updateMode(machine, entities.ModeStatic)
	}

	return nil
}

func (s *Service) startPollingInternal(machineID string, intervalMs int) {
	if intervalMs <= 0 {
		intervalMs = 1000
	}

	pollCtx, cancel := context.WithCancel(context.Background())
	s.pollingCancel.Store(machineID, cancel)

	go s.pollRoutine(pollCtx, machineID, time.Duration(intervalMs)*time.Millisecond)
}

func (s *Service) pollRoutine(ctx context.Context, machineID string, interval time.Duration) {
	log.Printf("Polling routine started for machine %s with interval %v", machineID, interval)

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Polling stopped for machine %s", machineID)
			return
		case <-timer.C:
			start := time.Now()

			// 1. Get or Restore Client
			client, err := s.getOrRestoreClient(ctx, machineID)
			if err != nil {
				log.Printf("Polling error for machine %s: %v. Status -> Reconnecting", machineID, err)
				if m, dbErr := s.repo.GetByID(machineID); dbErr == nil {
					s.updateStatus(m, entities.StatusReconnecting)
				}
				timer.Reset(5 * time.Second)
				continue
			}

			if m, dbErr := s.repo.GetByID(machineID); dbErr == nil && m.Status == entities.StatusReconnecting {
				s.updateStatus(m, entities.StatusConnected)
			}

			// 2. Execute Poll
			data, err := client.GetCurrentData()
			if err != nil {
				log.Printf("Error getting data from machine %s: %v", machineID, err)
				if m, dbErr := s.repo.GetByID(machineID); dbErr == nil {
					s.updateStatus(m, entities.StatusReconnecting)
				}
				s.clients.Delete(machineID)
			} else {
				// 3. Send to Kafka
				data.MachineID = machineID
				payload, err := json.Marshal(data)
				if err != nil {
					log.Printf("Failed to marshal polling data: %v", err)
				} else {
					if err := s.kafkaProducer.Send(context.Background(), []byte(machineID), payload); err != nil {
						log.Printf("Failed to send polling data to Kafka: %v", err)
					}
				}
			}

			elapsed := time.Since(start)
			nextWait := interval - elapsed
			if nextWait <= 0 {
				timer.Reset(0)
			} else {
				timer.Reset(nextWait)
			}
		}
	}
}

func (s *Service) getOrRestoreClient(ctx context.Context, id string) (*adapter.Client, error) {
	if val, ok := s.clients.Load(id); ok {
		return val.(*adapter.Client), nil
	}

	machine, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	ip, port, err := parseEndpoint(machine.Endpoint)
	if err != nil {
		return nil, err
	}

	cfg := &adapter.Config{
		IP:          ip,
		Port:        port,
		TimeoutMs:   int32(machine.Timeout),
		ModelSeries: machine.Series,
	}

	client, err := s.connectWithTimeout(cfg)
	if err != nil {
		return nil, err
	}

	s.clients.Store(id, client)
	return client, nil
}
