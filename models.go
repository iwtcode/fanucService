package fanucService

import "time"

// ConnectionRequest represents the payload to create a connection
type ConnectionRequest struct {
	Endpoint string `json:"endpoint" binding:"required"` // ip:port
	Timeout  int    `json:"timeout"`                     // ms, default 5000
	Model    string `json:"model"`                       // Human readable name, default "Unknown"
	Series   string `json:"series"`                      // "0i", "31i", default "Unknown"
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
	Timeout   int       `json:"timeout"`
	Model     string    `json:"model"`
	Series    string    `json:"series"`
	Interval  int       `json:"interval"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
