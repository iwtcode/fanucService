package fanuc

import (
	"context"
	"fmt"

	"github.com/iwtcode/fanucService/internal/domain/entities"
)

func (s *Service) GetControlProgram(ctx context.Context, id string) (string, error) {
	client, err := s.getOrRestoreClient(id)
	if err != nil {
		if m, dbErr := s.repo.GetByID(id); dbErr == nil {
			s.updateStatus(m, entities.StatusReconnecting)
		}
		return "", fmt.Errorf("machine unreachable: %w", err)
	}

	// Если клиент получен успешно, обновляем статус, если он был плохим
	if m, dbErr := s.repo.GetByID(id); dbErr == nil && m.Status == entities.StatusReconnecting {
		s.updateStatus(m, entities.StatusConnected)
	}

	// Запрашиваем программу у адаптера
	program, err := client.GetControlProgram()
	if err != nil {
		return "", fmt.Errorf("failed to download program: %w", err)
	}

	return program, nil
}
