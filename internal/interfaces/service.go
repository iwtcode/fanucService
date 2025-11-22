package interfaces

import (
	"context"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

// FanucService handles the actual communication with machines and state management
type FanucService interface {
	CreateConnection(ctx context.Context, req models.ConnectionRequest) (*entities.Machine, error)
	GetConnections(ctx context.Context) ([]entities.Machine, error)
	DeleteConnection(ctx context.Context, id string) error
	CheckConnection(ctx context.Context, id string) (*entities.Machine, error)
	RestoreConnections() error
}
