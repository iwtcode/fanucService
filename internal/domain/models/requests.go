package models

type ConnectionRequest struct {
	Endpoint string `json:"endpoint" binding:"required"` // ip:port
	Timeout  int    `json:"timeout"`                     // ms, default 5000
	Model    string `json:"model"`                       // Human readable name
	Series   string `json:"series"`                      // "0i", "31i"
}
