package fanuc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	adapter "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/services/kafka"
)

const (
	// HardConnectionTimeout sets a hard limit on any connection attempt
	HardConnectionTimeout = 5 * time.Second
	DefaultTimeout        = 5000
	DefaultUnknown        = "Unknown"
)

type Service struct {
	repo          interfaces.Repository
	kafkaProducer *kafka.Producer
	clients       sync.Map // map[string]*adapter.Client (Key: Machine ID)
	pollingCancel sync.Map // map[string]context.CancelFunc (Key: Machine ID)
}

// connectResult is used to pass results from goroutines
type connectResult struct {
	client *adapter.Client
	err    error
}

func NewService(repo interfaces.Repository, producer *kafka.Producer) interfaces.FanucService {
	return &Service{
		repo:          repo,
		kafkaProducer: producer,
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

// connectWithTimeout executes FOCAS connection in a separate goroutine with a hard timeout
func (s *Service) connectWithTimeout(cfg *adapter.Config) (*adapter.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), HardConnectionTimeout)
	defer cancel()

	resultCh := make(chan connectResult, 1)

	go func() {
		client, err := adapter.New(cfg)

		// If context expired, close connection immediately
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
