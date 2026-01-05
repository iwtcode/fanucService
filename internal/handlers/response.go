package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService/internal/domain/models"
)

func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Status: "ok",
		Data:   data,
	})
}

func RespondMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "ok",
		Message: message,
	})
}

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
