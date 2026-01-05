package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger возвращает middleware, который логирует HTTP запросы через logrus
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Старт таймера
		startTime := time.Now()

		// Обработка запроса
		c.Next()

		// Конец таймера
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Данные запроса
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		method := c.Request.Method
		path := c.Request.URL.Path
		dataLength := c.Writer.Size()

		if dataLength < 0 {
			dataLength = 0
		}

		entry := logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         clientIP,
			"latency_ms": latency.Milliseconds(),
			"user_agent": clientUserAgent,
			"data_len":   dataLength,
		})

		if len(c.Errors) > 0 {
			// Если есть ошибки в контексте Gin
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			if statusCode >= 500 {
				entry.Error("Internal Server Error")
			} else if statusCode >= 400 {
				entry.Warn("Client Error")
			} else {
				entry.Info("Request processed")
			}
		}
	}
}
