package fanucService

import "time"

// ConnectionRequest represents the payload to create a connection
type ConnectionRequest struct {
	Endpoint string `json:"endpoint" binding:"required"` // ip:port
	Series   string `json:"series" binding:"required"`   // "0i", "31i", etc
	Timeout  int    `json:"timeout"`                     // ms, default 5000
}

// ConnectionResponse represents a generic response wrapper
type ConnectionResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// MachineDTO represents the machine data sent to clients
type MachineDTO struct {
	ID        string    `json:"id"`
	Endpoint  string    `json:"endpoint"`
	Series    string    `json:"series"`
	Status    string    `json:"status"`
	Timeout   int       `json:"timeout"`
	Interval  int       `json:"interval"`
	CreatedAt time.Time `json:"created_at"`
}
