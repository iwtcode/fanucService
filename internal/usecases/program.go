package usecases

import (
	"context"

	"github.com/iwtcode/fanucService/internal/interfaces"
)

type programUsecase struct {
	service interfaces.FanucService
}

func NewProgramUsecase(service interfaces.FanucService) interfaces.ProgramUsecase {
	return &programUsecase{service: service}
}

func (u *programUsecase) GetProgram(ctx context.Context, id string) (string, error) {
	return u.service.GetControlProgram(ctx, id)
}
