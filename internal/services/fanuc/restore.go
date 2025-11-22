package fanuc

import (
	"time"

	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/domain/entities"
)

func (s *Service) RestoreConnections() error {
	machines, err := s.repo.GetAll()
	if err != nil {
		return err
	}

	for _, m := range machines {
		go func(machine entities.Machine) {
			ip, port, err := parseEndpoint(machine.Endpoint)
			if err != nil {
				return
			}

			cfg := &adapter.Config{
				IP:          ip,
				Port:        port,
				TimeoutMs:   int32(machine.Timeout),
				ModelSeries: machine.Series,
			}

			client, err := s.connectWithTimeout(cfg)
			if err == nil {
				s.clients.Store(machine.ID, client)
				if machine.Status != entities.StatusConnected {
					machine.Status = entities.StatusConnected
					machine.UpdatedAt = time.Now()
					_ = s.repo.Update(&machine)
				}
			} else {
				if machine.Status != entities.StatusReconnecting {
					machine.Status = entities.StatusReconnecting
					machine.UpdatedAt = time.Now()
					_ = s.repo.Update(&machine)
				}
			}
		}(m)
	}
	return nil
}
