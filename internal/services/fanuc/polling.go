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

	// 2. Ensure connection exists and machine is reachable
	// This also fetches the machine entity from DB
	machine, err := s.CheckConnection(ctx, machineID)
	if err != nil {
		return fmt.Errorf("cannot start polling, machine unreachable: %w", err)
	}

	// 3. Get client from pool
	val, ok := s.clients.Load(machineID)
	if !ok {
		return fmt.Errorf("client not found in pool for machine %s", machineID)
	}
	client := val.(*adapter.Client)

	// 4. Update Status in DB to 'polled'
	// Note: We use the helper method updateStatus defined in connection.go (same package)
	s.updateStatus(machine, entities.StatusPolled)

	// 5. Create cancellation context
	pollCtx, cancel := context.WithCancel(context.Background())
	s.pollingCancel.Store(machineID, cancel)

	// 6. Start background polling routine
	go s.pollRoutine(pollCtx, machineID, client, time.Duration(intervalMs)*time.Millisecond)

	return nil
}

func (s *Service) StopPolling(ctx context.Context, machineID string) error {
	val, ok := s.pollingCancel.Load(machineID)
	if !ok {
		return fmt.Errorf("polling not active for machine %s", machineID)
	}

	// Stop the goroutine
	cancel := val.(context.CancelFunc)
	cancel()
	s.pollingCancel.Delete(machineID)

	// Update Status in DB back to 'connected'
	machine, err := s.repo.GetByID(machineID)
	if err == nil {
		s.updateStatus(machine, entities.StatusConnected)
	} else {
		log.Printf("Warning: failed to fetch machine %s to update status on stop polling: %v", machineID, err)
	}

	return nil
}

func (s *Service) pollRoutine(ctx context.Context, machineID string, client *adapter.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Polling started for machine %s with interval %v", machineID, interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Polling stopped for machine %s", machineID)
			return
		case <-ticker.C:
			// Perform polling
			data, err := client.GetCurrentData()
			if err != nil {
				log.Printf("Error polling machine %s: %v", machineID, err)
				// On error, we might consider stopping, but usually we retry.
				// If the machine becomes completely unreachable, CheckConnection or a background health check
				// should eventually handle status updates, but here we just log.
				continue
			}

			// Enrich data with ID
			data.MachineID = machineID

			// Send to Kafka
			payload, err := json.Marshal(data)
			if err != nil {
				log.Printf("Failed to marshal polling data: %v", err)
				continue
			}

			// Send asynchronously
			err = s.kafkaProducer.Send(context.Background(), []byte(machineID), payload)
			if err != nil {
				log.Printf("Failed to send polling data to Kafka: %v", err)
			}
		}
	}
}
