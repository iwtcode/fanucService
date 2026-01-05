package fanuc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/services/kafka"
	"github.com/sirupsen/logrus"
)

const (
	HardConnectionTimeout = 5 * time.Second
	DefaultTimeout        = 5000
	DefaultUnknown        = "Unknown"
)

type Service struct {
	cfg           *fanucService.Config
	repo          interfaces.Repository
	kafkaProducer *kafka.Producer
	logger        *logrus.Logger
	clients       sync.Map
	pollingCancel sync.Map
}

type connectResult struct {
	client *adapter.Client
	err    error
}

func NewService(cfg *fanucService.Config, repo interfaces.Repository, producer *kafka.Producer, logger *logrus.Logger) interfaces.FanucService {
	return &Service{
		cfg:           cfg,
		repo:          repo,
		kafkaProducer: producer,
		logger:        logger,
	}
}

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

func (s *Service) connectWithTimeout(cfg *adapter.Config) (*adapter.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), HardConnectionTimeout)
	defer cancel()

	resultCh := make(chan connectResult, 1)

	go func() {
		client, err := adapter.New(cfg)

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
