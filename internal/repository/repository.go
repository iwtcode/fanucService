package repository

import (
	"fmt"
	"log"

	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/interfaces"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(cfg *fanucService.Config) (interfaces.Repository, error) {
	// 1. Connect to default 'postgres' database to check/create target DB
	dsnRoot := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password)

	rootDB, err := gorm.Open(postgres.Open(dsnRoot), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to root postgres db: %w", err)
	}

	var exists bool
	checkQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", cfg.Database.Name)
	rootDB.Raw(checkQuery).Scan(&exists)

	if !exists {
		log.Printf("Database %s does not exist. Creating...", cfg.Database.Name)
		if err := rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database.Name)).Error; err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}
	sqlDB, _ := rootDB.DB()
	sqlDB.Close()

	// 2. Connect to actual database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to application db: %w", err)
	}

	// Auto Migrate
	if err := db.AutoMigrate(&entities.Machine{}); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &postgresRepository{db: db}, nil
}

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
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) GetByEndpoint(endpoint string) (*entities.Machine, error) {
	var m entities.Machine
	err := r.db.First(&m, "endpoint = ?", endpoint).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) GetAll() ([]entities.Machine, error) {
	var list []entities.Machine
	err := r.db.Find(&list).Error
	return list, err
}
