package interfaces

import "github.com/iwtcode/fanucService/internal/domain/entities"

type Repository interface {
	Create(machine *entities.Machine) error
	Update(machine *entities.Machine) error
	Delete(id string) error
	GetByID(id string) (*entities.Machine, error)
	GetByEndpoint(endpoint string) (*entities.Machine, error)
	GetAll() ([]entities.Machine, error)
}
