package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService"
)

func Auth(cfg *fanucService.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}

		if key != cfg.App.APIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "unauthorized"})
			return
		}
		c.Next()
	}
}
