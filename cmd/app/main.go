package main

import (
	"github.com/iwtcode/fanucService/internal/app"
)

// @title Fanuc Service API
// @version 1.0
// @description Service for managing Fanuc CNC connections and data polling
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @BasePath /
func main() {
	app.New().Run()
}
