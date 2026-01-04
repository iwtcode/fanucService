package fanuc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

func (s *Service) CreateConnection(ctx context.Context, req models.ConnectionRequest) (*entities.Machine, error) {
	existing, _ := s.repo.GetByEndpoint(req.Endpoint)
	if existing != nil {
		return nil, fmt.Errorf("connection to %s already exists with ID %s", req.Endpoint, existing.ID)
	}

	series := req.Series
	if series == "" {
		series = DefaultUnknown
	}

	model := req.Model
	if model == "" {
		model = DefaultUnknown
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if timeout > int(HardConnectionTimeout.Milliseconds()) {
		timeout = int(HardConnectionTimeout.Milliseconds())
	}

	ip, port, err := parseEndpoint(req.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint format: %w", err)
	}

	adapterCfg := &adapter.Config{
		IP:          ip,
		Port:        port,
		TimeoutMs:   int32(timeout),
		ModelSeries: series,
		LogLevel:    s.cfg.Adapter.LogLevel,
	}

	client, err := s.connectWithTimeout(adapterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to machine: %w", err)
	}

	machine := &entities.Machine{
		ID:        uuid.New().String(),
		Endpoint:  req.Endpoint,
		Timeout:   timeout,
		Model:     model,
		Series:    series,
		Status:    entities.StatusConnected,
		Mode:      entities.ModeStatic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(machine); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to save machine to db: %w", err)
	}

	s.clients.Store(machine.ID, client)

	return machine, nil
}

func (s *Service) GetConnections(ctx context.Context) ([]entities.Machine, error) {
	machines, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	if len(machines) == 0 {
		return machines, nil
	}

	var wg sync.WaitGroup
	results := make([]entities.Machine, len(machines))

	for i, m := range machines {
		wg.Add(1)
		go func(index int, id string, original entities.Machine) {
			defer wg.Done()
			updatedMachine, _ := s.CheckConnection(ctx, id)
			if updatedMachine != nil {
				results[index] = *updatedMachine
			} else {
				results[index] = original
			}
		}(i, m.ID, m)
	}

	wg.Wait()
	return results, nil
}

func (s *Service) DeleteConnection(ctx context.Context, id string) error {
	if val, ok := s.pollingCancel.Load(id); ok {
		cancel := val.(context.CancelFunc)
		cancel()
		s.pollingCancel.Delete(id)
	}

	if val, ok := s.clients.Load(id); ok {
		client := val.(*adapter.Client)
		client.Close()
		s.clients.Delete(id)
	}
	return s.repo.Delete(id)
}

func (s *Service) CheckConnection(ctx context.Context, id string) (*entities.Machine, error) {
	machine, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	var client *adapter.Client
	var inPool bool

	if val, found := s.clients.Load(id); found {
		client = val.(*adapter.Client)
		inPool = true
	}

	if !inPool {
		ip, port, _ := parseEndpoint(machine.Endpoint)
		cfg := &adapter.Config{
			IP:          ip,
			Port:        port,
			TimeoutMs:   int32(machine.Timeout),
			ModelSeries: machine.Series,
			LogLevel:    s.cfg.Adapter.LogLevel,
		}

		client, err = s.connectWithTimeout(cfg)
		if err != nil {
			s.updateStatus(machine, entities.StatusReconnecting)
			return machine, fmt.Errorf("machine unreachable: %w", err)
		}
		s.clients.Store(id, client)
	}

	checkErrChan := make(chan error, 1)
	go func() {
		_, err := client.GetMachineState()
		checkErrChan <- err
	}()

	select {
	case err := <-checkErrChan:
		if err != nil {
			client.Close()
			s.clients.Delete(id)
			s.updateStatus(machine, entities.StatusReconnecting)
			return machine, fmt.Errorf("health check failed: %w", err)
		}
	case <-time.After(HardConnectionTimeout):
		s.updateStatus(machine, entities.StatusReconnecting)
		return machine, fmt.Errorf("health check timed out")
	}

	s.updateStatus(machine, entities.StatusConnected)

	return machine, nil
}

func (s *Service) updateStatus(m *entities.Machine, status string) {
	if m.Status != status {
		m.Status = status
		m.UpdatedAt = time.Now()
		_ = s.repo.Update(m)
	}
}

func (s *Service) updateMode(m *entities.Machine, mode string) {
	if m.Mode != mode {
		m.Mode = mode
		m.UpdatedAt = time.Now()
		_ = s.repo.Update(m)
	}
}

func (s *Service) updateInterval(m *entities.Machine, interval int) {
	if m.Interval != interval {
		m.Interval = interval
		m.UpdatedAt = time.Now()
		_ = s.repo.Update(m)
	}
}
