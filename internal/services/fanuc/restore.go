package fanuc

import (
	"log"
	"time"

	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/domain/entities"
)

func (s *Service) RestoreConnections() error {
	machines, err := s.repo.GetAll()
	if err != nil {
		return err
	}

	log.Printf("Restoring connections for %d machines...", len(machines))

	for _, m := range machines {
		go s.restoreOne(m)
	}
	return nil
}

func (s *Service) restoreOne(machine entities.Machine) {
	ip, port, err := parseEndpoint(machine.Endpoint)
	if err != nil {
		log.Printf("Invalid endpoint for machine %s: %v", machine.ID, err)
		return
	}

	cfg := &adapter.Config{
		IP:          ip,
		Port:        port,
		TimeoutMs:   int32(machine.Timeout),
		ModelSeries: machine.Series,
	}

	// Attempt connection
	client, err := s.connectWithTimeout(cfg)

	if err == nil {
		// Connection Successful
		s.clients.Store(machine.ID, client)
		log.Printf("Restored connection to %s", machine.Endpoint)

		if machine.Status == entities.StatusPolled {
			// Resume polling immediately
			log.Printf("Resuming polling for machine %s", machine.ID)
			s.startPollingInternal(machine.ID, client, machine.Interval)
		} else {
			// Ensure status is consistent (fix if it was 'reconnecting' in old db)
			s.updateStatus(&machine, entities.StatusConnected)
		}
	} else {
		// Connection Failed
		log.Printf("Failed to restore connection to %s: %v", machine.Endpoint, err)

		// If it was supposed to be polling, we MUST eventually start polling.
		// Start a background retry loop.
		if machine.Status == entities.StatusPolled {
			log.Printf("Machine %s is in Polled status but offline. Starting background retry routine...", machine.ID)
			go s.restorePollingInBackground(machine, cfg)
		}

		// Note: We do NOT change status to 'reconnecting'.
		// If it was 'connected', it stays 'connected' (but offline).
		// If it was 'polled', it stays 'polled' (waiting for connection).
	}
}

// restorePollingInBackground keeps trying to connect to a 'polled' machine until successful, then starts polling.
func (s *Service) restorePollingInBackground(machine entities.Machine, cfg *adapter.Config) {
	ticker := time.NewTicker(10 * time.Second) // Retry every 10 seconds
	defer ticker.Stop()

	for range ticker.C {
		// Check if polling was cancelled (e.g. user called StopPolling while we were retrying)
		// Since we haven't started the pollRoutine yet, we check the DB or a simplified check.
		// However, simpler: if user calls StopPolling, they remove 'polled' status from DB.
		// Let's check DB fresh status to be sure we should still be trying.
		currentMachine, err := s.repo.GetByID(machine.ID)
		if err != nil || currentMachine.Status != entities.StatusPolled {
			log.Printf("Stopping retry loop for machine %s (status changed or removed)", machine.ID)
			return
		}

		log.Printf("Retrying connection for polled machine %s...", machine.Endpoint)
		client, err := s.connectWithTimeout(cfg)
		if err == nil {
			log.Printf("Connection established for machine %s! Starting polling.", machine.Endpoint)
			s.clients.Store(machine.ID, client)
			s.startPollingInternal(machine.ID, client, machine.Interval)
			return // Exit retry loop
		}
		// Failure logic: continue loop
	}
}
