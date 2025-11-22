package entities

import (
	"time"
)

const (
	// Status - физическое состояние подключения
	StatusConnected    = "connected"
	StatusReconnecting = "reconnecting"

	// Mode - режим работы сервиса по отношению к станку
	ModeStatic  = "static"
	ModePolling = "polling"
)

type Machine struct {
	ID       string `gorm:"primaryKey;type:uuid" json:"id"`       // uuid
	Endpoint string `gorm:"uniqueIndex;not null" json:"endpoint"` // ip:port
	Timeout  int    `json:"timeout"`                              // таймаут в мс
	Model    string `json:"model"`                                // Human readable model name
	Series   string `json:"series"`                               // "0i", "31i"
	Interval int    `json:"interval"`                             // Интервал опроса в мс

	Status string `gorm:"not null;default:'reconnecting'" json:"status"` // connected / reconnecting
	Mode   string `gorm:"not null;default:'static'" json:"mode"`         // static / polling

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
