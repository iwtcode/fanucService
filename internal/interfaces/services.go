package interfaces

import (
	"time"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

// FanucService - это агрегирующий интерфейс для всей бизнес-логики.
type FanucService interface {
	ConnectionManager
	PollingManager
}

// ConnectionManager определяет контракт для управления пулом подключений.
type ConnectionManager interface {
	CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error)
	RestoreConnection(machine entities.FanucMachine) (*models.ConnectionInfo, error)
	GetConnection(sessionID string) (*models.ConnectionInfo, bool)
	GetAllConnections() []*models.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*models.ConnectionInfo, error)
}

// PollingManager определяет контракт для сервиса, опрашивающего эндпоинты.
type PollingManager interface {
	StartPolling(conn *models.ConnectionInfo, interval time.Duration) error
	StopPolling(sessionID string) error
	IsPollingActive(sessionID string) bool
}
