package usecases

import "github.com/iwtcode/fanucService/internal/interfaces"

type restoreUsecase struct {
	service interfaces.FanucService
}

func NewRestoreUsecase(service interfaces.FanucService) interfaces.RestoreUsecase {
	return &restoreUsecase{service: service}
}

func (u *restoreUsecase) Restore() {
	go u.service.RestoreConnections()
}
