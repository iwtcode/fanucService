package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

// RespondSuccess sends a standardized 200 OK response
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Status: "ok",
		Data:   data,
	})
}

// RespondError sends a standardized error response
func RespondError(c *gin.Context, code int, message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	c.JSON(code, models.APIResponse{
		Status:  "error",
		Message: message,
		Data:    d,
	})
}
