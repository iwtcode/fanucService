package fanuc

import (
	"log"

	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/domain/entities"
)

func (s *Service) RestoreConnections() error {
	machines, err := s.repo.GetAll()
	if err != nil {
		return err
	}

	log.Printf("Restoring state for %d machines...", len(machines))

	for _, m := range machines {
		if m.Mode == entities.ModePolling {
			log.Printf("Machine %s is in Polling mode. Starting polling routine...", m.ID)
			s.startPollingInternal(m.ID, m.Interval)
			continue
		}

		go s.checkOneOnce(m)
	}
	return nil
}

func (s *Service) checkOneOnce(machine entities.Machine) {
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
		log.Printf("Restored connection to %s (Static mode)", machine.Endpoint)
		s.updateStatus(&machine, entities.StatusConnected)
	} else {
		// Connection Failed
		log.Printf("Machine %s (Static mode) is unreachable: %v", machine.Endpoint, err)
		s.updateStatus(&machine, entities.StatusReconnecting)
	}
}
