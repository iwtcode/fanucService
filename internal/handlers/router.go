package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService"
	_ "github.com/iwtcode/fanucService/docs" // Import docs
	"github.com/iwtcode/fanucService/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg *fanucService.Config, connHandler *ConnectionHandler) *gin.Engine {
	gin.SetMode(cfg.App.GinMode)
	r := gin.Default()

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Group
	v1 := r.Group("/api/v1")
	v1.Use(middleware.Auth(cfg))
	{
		connect := v1.Group("/connect")
		{
			connect.POST("", connHandler.Create)
			connect.GET("", connHandler.Get) // Обрабатывает и список, и проверку по ID
			connect.DELETE("", connHandler.Delete)
		}
	}

	return r
}
