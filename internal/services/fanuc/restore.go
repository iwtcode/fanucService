package fanuc

import (
	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/domain/entities"
)

func (s *Service) RestoreConnections() error {
	machines, err := s.repo.GetAll()
	if err != nil {
		return err
	}

	s.logger.Infof("Restoring state for %d machines...", len(machines))
	go func() {
		for _, m := range machines {
			if m.Mode == entities.ModePolling {
				s.logger.Infof("Machine %s is in Polling mode. Starting polling routine...", m.ID)
				s.startPollingInternal(m.ID, m.Interval)
				continue
			}
			s.checkOneOnce(m)
		}
	}()

	return nil
}

func (s *Service) checkOneOnce(machine entities.Machine) {
	ip, port, err := parseEndpoint(machine.Endpoint)
	if err != nil {
		s.logger.Errorf("Invalid endpoint for machine %s: %v", machine.ID, err)
		return
	}

	cfg := &adapter.Config{
		IP:          ip,
		Port:        port,
		TimeoutMs:   int32(machine.Timeout),
		ModelSeries: machine.Series,
		LogLevel:    s.cfg.Logger.AdapterLevel,
	}

	client, err := s.connectWithTimeout(cfg)

	if err == nil {
		s.clients.Store(machine.ID, client)
		s.logger.Infof("Restored connection to %s (Static mode)", machine.Endpoint)
		s.updateStatus(&machine, entities.StatusConnected)
	} else {
		s.logger.Warnf("Machine %s (Static mode) is unreachable: %v", machine.Endpoint, err)
		s.updateStatus(&machine, entities.StatusReconnecting)
	}
}
