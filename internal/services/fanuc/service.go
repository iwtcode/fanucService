package fanuc

import (
	"sync"

	"github.com/iwtcode/fanucService/internal/interfaces"
)

type Service struct {
	repo    interfaces.Repository
	clients sync.Map // map[string]*adapter.Client (Key: Machine ID)
}

func NewService(repo interfaces.Repository) interfaces.FanucService {
	return &Service{
		repo: repo,
	}
}
