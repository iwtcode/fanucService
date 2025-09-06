package interfaces

import (
	"time"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

// Usecases - это агрегирующий интерфейс для всех use cases
type Usecases interface {
	CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error)
	RestoreConnection(machine entities.FanucMachine) (*models.ConnectionInfo, error)
	GetAllConnections() []*models.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*models.ConnectionInfo, error)
	StartPolling(sessionID string, interval time.Duration) error
	StopPolling(sessionID string) error
}
