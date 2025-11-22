package entities

import (
	"time"
)

const (
	StatusConnected    = "connected"
	StatusPolled       = "polled"
	StatusReconnecting = "reconnecting" // пытаемся переподключиться к станку / станок пока не доступен
)

type Machine struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`       // uuid
	Endpoint  string    `gorm:"uniqueIndex;not null" json:"endpoint"` // ip:port
	Timeout   int       `json:"timeout"`                              // таймаут в мс
	Model     string    `json:"model"`                                // Human readable model name
	Series    string    `json:"series"`                               // "0i", "31i"
	Interval  int       `json:"interval"`                             // Интервал опроса в мс
	Status    string    `gorm:"not null" json:"status"`               // connected / polled / reconnecting
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
