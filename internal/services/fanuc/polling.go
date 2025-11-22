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

	// 2. Ensure connection exists and machine is reachable (REAL CHECK)
	// CheckConnection выполняет реальный запрос GetMachineState.
	// Если станок недоступен, вернется ошибка.
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

	// 4. Update Status & Interval in DB
	s.updateInterval(machine, intervalMs)
	s.updateStatus(machine, entities.StatusPolled)

	// 5. Start internal polling
	s.startPollingInternal(machineID, client, intervalMs)

	return nil
}

// startPollingInternal starts the routine without DB updates or checks (used for restore)
func (s *Service) startPollingInternal(machineID string, client *adapter.Client, intervalMs int) {
	if intervalMs <= 0 {
		intervalMs = 1000 // safe fallback
	}

	pollCtx, cancel := context.WithCancel(context.Background())
	s.pollingCancel.Store(machineID, cancel)

	go s.pollRoutine(pollCtx, machineID, client, time.Duration(intervalMs)*time.Millisecond)
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
	}

	return nil
}

func (s *Service) pollRoutine(ctx context.Context, machineID string, client *adapter.Client, interval time.Duration) {
	// Первый запуск - через интервал
	timer := time.NewTimer(interval)
	defer timer.Stop()

	log.Printf("Polling started for machine %s with interval %v", machineID, interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Polling stopped for machine %s", machineID)
			return
		case <-timer.C:
			// 1. Засекаем время начала работы
			start := time.Now()

			// 2. Выполняем опрос
			data, err := client.GetCurrentData()
			if err != nil {
				log.Printf("Error polling machine %s: %v", machineID, err)
			} else {
				// 3. Отправляем данные в Kafka (только если опрос успешен)
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

			// 4. Вычисляем затраченное время
			elapsed := time.Since(start)

			// 5. Вычисляем время до следующего запуска
			// Цель: следующий запуск должен быть ровно через interval после НАЧАЛА текущего.
			nextWait := interval - elapsed

			if nextWait <= 0 {
				// Если работа заняла больше времени, чем интервал, запускаем следующую итерацию немедленно
				timer.Reset(0)
			} else {
				// Ждем оставшееся время
				timer.Reset(nextWait)
			}
		}
	}
}
