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
	Series    string    `gorm:"not null" json:"series"`               // "0i", "31i"
	Status    string    `gorm:"not null" json:"status"`               // connected / polled / reconnecting (disconnected = delete /connect = удален из БД, статус не нужен)
	Interval  int       `json:"interval"`                             // Интервал опроса в мс
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
