package usecases

import (
	"context"

	"github.com/iwtcode/fanucService/internal/domain/models"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

type pollingUsecase struct {
	service interfaces.FanucService
}

func NewPollingUsecase(service interfaces.FanucService) interfaces.PollingUsecase {
	return &pollingUsecase{service: service}
}

func (u *pollingUsecase) Start(ctx context.Context, req models.StartPollingRequest) error {
	if req.Interval <= 0 {
		req.Interval = 10000
	}

	return u.service.StartPolling(ctx, req.ID, req.Interval)
}

func (u *pollingUsecase) Stop(ctx context.Context, req models.StopPollingRequest) error {
	return u.service.StopPolling(ctx, req.ID)
}
