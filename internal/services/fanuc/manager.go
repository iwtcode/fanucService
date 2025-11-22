package fanuc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

const (
	// HardConnectionTimeout задает жесткий лимит времени на любую попытку подключения.
	HardConnectionTimeout = 5 * time.Second
	DefaultTimeout        = 5000
	DefaultUnknown        = "Unknown"
)

type Service struct {
	repo    interfaces.Repository
	clients sync.Map // map[string]*adapter.Client (Key: Machine ID)
}

func NewService(repo interfaces.Repository) interfaces.FanucService {
	return &Service{
		repo: repo,
	}
}

// Helper to parse "ip:port"
func parseEndpoint(endpoint string) (string, uint16, error) {
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", 0, err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, err
	}
	return host, uint16(port), nil
}

// connectResult используется для передачи результата из горутины
type connectResult struct {
	client *adapter.Client
	err    error
}

// connectWithTimeout выполняет подключение к FOCAS в отдельной горутине с жестким таймаутом.
func (s *Service) connectWithTimeout(cfg *adapter.Config) (*adapter.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), HardConnectionTimeout)
	defer cancel()

	resultCh := make(chan connectResult, 1)

	go func() {
		client, err := adapter.New(cfg)

		// Если контекст истек, закрываем соединение сразу
		if ctx.Err() != nil {
			if client != nil {
				client.Close()
			}
			return
		}

		resultCh <- connectResult{client: client, err: err}
	}()

	select {
	case res := <-resultCh:
		if res.err != nil {
			return nil, res.err
		}
		return res.client, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("hard timeout: failed to connect within %v", HardConnectionTimeout)
	}
}

func (s *Service) CreateConnection(ctx context.Context, req fanucService.ConnectionRequest) (*entities.Machine, error) {
	// 1. Проверка дубликатов
	existing, _ := s.repo.GetByEndpoint(req.Endpoint)
	if existing != nil {
		return nil, fmt.Errorf("connection to %s already exists with ID %s", req.Endpoint, existing.ID)
	}

	// 2. Применение значений по умолчанию
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
	// Ограничиваем таймаут сверху константой хард-таймаута (5000мс), чтобы не блокировать горутины надолго
	if timeout > int(HardConnectionTimeout.Milliseconds()) {
		timeout = int(HardConnectionTimeout.Milliseconds())
	}

	// 3. Парсинг адреса
	ip, port, err := parseEndpoint(req.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint format: %w", err)
	}

	adapterCfg := &adapter.Config{
		IP:          ip,
		Port:        port,
		TimeoutMs:   int32(timeout),
		ModelSeries: series,
		LogPath:     "./focas.log",
	}

	// 4. Подключение
	client, err := s.connectWithTimeout(adapterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to machine: %w", err)
	}

	// 5. Сохранение в БД
	machine := &entities.Machine{
		ID:        uuid.New().String(),
		Endpoint:  req.Endpoint,
		Timeout:   timeout,
		Model:     model,
		Series:    series,
		Status:    entities.StatusConnected,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(machine); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to save machine to db: %w", err)
	}

	// 6. Сохранение в пул
	s.clients.Store(machine.ID, client)

	return machine, nil
}

// GetConnections возвращает список станков, предварительно проверив их состояние параллельно.
func (s *Service) GetConnections(ctx context.Context) ([]entities.Machine, error) {
	// 1. Получаем список из БД
	machines, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	if len(machines) == 0 {
		return machines, nil
	}

	// 2. Запускаем параллельную проверку
	var wg sync.WaitGroup
	results := make([]entities.Machine, len(machines))

	for i, m := range machines {
		wg.Add(1)
		go func(index int, id string, original entities.Machine) {
			defer wg.Done()

			updatedMachine, err := s.CheckConnection(ctx, id)

			if updatedMachine != nil {
				results[index] = *updatedMachine
			} else {
				if err != nil {
					fmt.Printf("Error checking machine %s: %v\n", id, err)
				}
				results[index] = original
			}
		}(i, m.ID, m)
	}

	wg.Wait()

	return results, nil
}

func (s *Service) DeleteConnection(ctx context.Context, id string) error {
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

	// 1. Проверяем пул
	if val, found := s.clients.Load(id); found {
		client = val.(*adapter.Client)
		inPool = true
	}

	// 2. Восстановление соединения, если нет в пуле
	if !inPool {
		ip, port, _ := parseEndpoint(machine.Endpoint)
		cfg := &adapter.Config{
			IP:          ip,
			Port:        port,
			TimeoutMs:   int32(machine.Timeout),
			ModelSeries: machine.Series,
		}

		client, err = s.connectWithTimeout(cfg)
		if err != nil {
			s.updateStatus(machine, entities.StatusReconnecting)
			return machine, fmt.Errorf("machine unreachable: %w", err)
		}
		s.clients.Store(id, client)
	}

	// 3. Проверка здоровья.
	checkErrChan := make(chan error, 1)
	go func() {
		// GetMachineState делает вызов cnc_statinfo, который идет по сети.
		_, err := client.GetMachineState()
		checkErrChan <- err
	}()

	select {
	case err := <-checkErrChan:
		if err != nil {
			// Ошибка сети (сокет закрыт, таймаут внутри библиотеки и т.д.)
			client.Close()
			s.clients.Delete(id)
			s.updateStatus(machine, entities.StatusReconnecting)
			return machine, fmt.Errorf("health check failed: %w", err)
		}
	case <-time.After(HardConnectionTimeout):
		// Наш жесткий таймаут сработал раньше, чем ответила библиотека
		s.updateStatus(machine, entities.StatusReconnecting)
		// Не удаляем клиента здесь жестко, возможно он "отвиснет", но статус обновляем
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
