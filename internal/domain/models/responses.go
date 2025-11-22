package models

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type MachineResponse struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
}
