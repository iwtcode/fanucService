package repository

import (
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

func (r *postgresRepository) Create(machine *entities.Machine) error {
	return r.db.Create(machine).Error
}

func (r *postgresRepository) Update(machine *entities.Machine) error {
	return r.db.Save(machine).Error
}

func (r *postgresRepository) Delete(id string) error {
	return r.db.Delete(&entities.Machine{}, "id = ?", id).Error
}

func (r *postgresRepository) GetByID(id string) (*entities.Machine, error) {
	var m entities.Machine
	err := r.db.First(&m, "id = ?", id).Error
	if err != nil {
		return nil, models.ErrNotFound
	}
	return &m, nil
}

func (r *postgresRepository) GetByEndpoint(endpoint string) (*entities.Machine, error) {
	var m entities.Machine
	err := r.db.First(&m, "endpoint = ?", endpoint).Error
	if err != nil {
		return nil, models.ErrNotFound
	}
	return &m, nil
}

func (r *postgresRepository) GetAll() ([]entities.Machine, error) {
	var list []entities.Machine
	err := r.db.Find(&list).Error
	return list, err
}
