package entities

import "time"

const (
	StatusConnected = "connected"
	StatusPolled    = "polled"
)

type FanucMachine struct {
	SessionID   string    `gorm:"primaryKey;not null" json:"session_id"`
	EndpointURL string    `gorm:"not null;unique" json:"endpoint_url"` // IP:PORT
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `gorm:"not null" json:"status"` // connected / polled
	Interval    int       `json:"interval"`               // Интервал опроса в мс
}
