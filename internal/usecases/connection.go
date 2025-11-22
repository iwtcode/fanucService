package usecases

import (
	"context"

	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

type connectionUsecase struct {
	service interfaces.FanucService
}

func NewConnectionUsecase(service interfaces.FanucService) interfaces.ConnectionUsecase {
	return &connectionUsecase{service: service}
}

func (u *connectionUsecase) Create(ctx context.Context, req fanucService.ConnectionRequest) (*entities.Machine, error) {
	return u.service.CreateConnection(ctx, req)
}

func (u *connectionUsecase) List(ctx context.Context) ([]entities.Machine, error) {
	return u.service.GetConnections(ctx)
}

func (u *connectionUsecase) Delete(ctx context.Context, id string) error {
	return u.service.DeleteConnection(ctx, id)
}

func (u *connectionUsecase) Check(ctx context.Context, id string) (*entities.Machine, error) {
	return u.service.CheckConnection(ctx, id)
}
