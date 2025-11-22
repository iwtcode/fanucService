package entities

import (
	"time"
)

const (
	StatusConnected = "connected"
	StatusPolled    = "polled"
	// StatusReconnecting удален по требованию: станок должен оставаться в статусе connected/polled даже при временной недоступности
)

type Machine struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`       // uuid
	Endpoint  string    `gorm:"uniqueIndex;not null" json:"endpoint"` // ip:port
	Timeout   int       `json:"timeout"`                              // таймаут в мс
	Model     string    `json:"model"`                                // Human readable model name
	Series    string    `json:"series"`                               // "0i", "31i"
	Interval  int       `json:"interval"`                             // Интервал опроса в мс
	Status    string    `gorm:"not null" json:"status"`               // connected / polled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
