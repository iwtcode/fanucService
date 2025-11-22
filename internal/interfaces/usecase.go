package interfaces

import (
	"context"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

type ConnectionUsecase interface {
	Create(ctx context.Context, req models.ConnectionRequest) (*entities.Machine, error)
	List(ctx context.Context) ([]entities.Machine, error)
	Delete(ctx context.Context, id string) error
	Check(ctx context.Context, id string) (*entities.Machine, error)
}

type RestoreUsecase interface {
	Restore()
}
