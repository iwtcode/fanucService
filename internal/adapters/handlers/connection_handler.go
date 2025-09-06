package handlers

import (
	"net/http"

	"github.com/iwtcode/fanucService/internal/domain/models"

	"github.com/gin-gonic/gin"
)

// CreateConnection создает новое подключение к Fanuc станку.
// @Summary Создать подключение
// @Description Создает новое подключение к станку по его IP-адресу и порту.
// @Tags Connection
// @Accept json
// @Produce json
// @Param input body models.ConnectionRequest true "Данные для подключения (e.g., '192.168.1.10:8193')"
// @Success 200 {object} models.CreateConnectionResponse "Успешное создание подключения"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера или станок недоступен"
// @Router /connect [post]
func (h *Handler) CreateConnection(c *gin.Context) {
	var req models.ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, err, "Invalid request payload")
		return
	}

	h.logger.Info("Attempting to create a new connection", "endpoint", req.EndpointURL)

	connInfo, err := h.usecase.CreateConnection(req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.logger.Info("Successfully created connection", "sessionID", connInfo.SessionID)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "connection_info": connInfo})
}

// GetConnections возвращает список всех активных подключений.
// @Summary Получить список подключений
// @Description Возвращает текущий пул активных Fanuc подключений.
// @Tags Connection
// @Produce json
// @Success 200 {object} models.GetConnectionsResponse "Список активных подключений"
// @Router /connect [get]
func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.usecase.GetAllConnections()
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"pool_size":   len(connections),
		"connections": connections,
	})
}

// DeleteConnection удаляет подключение по SessionID.
// @Summary Удалить подключение
// @Description Удаляет подключение из пула, останавливает опрос и удаляет запись из БД.
// @Tags Connection
// @Accept json
// @Produce json
// @Param input body models.SessionRequest true "ID сессии для удаления"
// @Success 200 {object} models.MessageResponse "Сообщение об успешном удалении"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 404 {object} models.ErrorResponse "Подключение не найдено"
// @Router /connect [delete]
func (h *Handler) DeleteConnection(c *gin.Context) {
	var req models.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, err, "Missing or invalid SessionID")
		return
	}

	h.logger.Info("Attempting to delete connection", "sessionID", req.SessionID)

	if err := h.usecase.DeleteConnection(req.SessionID); err != nil {
		h.NotFound(c, err)
		return
	}

	h.logger.Info("Successfully deleted connection", "sessionID", req.SessionID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Session " + req.SessionID + " disconnected successfully",
	})
}

// CheckConnection проверяет состояние подключения по SessionID.
// @Summary Проверить состояние подключения
// @Description Проверяет доступность эндпоинта, связанного с указанным SessionID.
// @Tags Connection
// @Accept json
// @Produce json
// @Param input body models.SessionRequest true "ID сессии для проверки"
// @Success 200 {object} models.CheckConnectionResponse "Статус 'healthy' или 'unhealthy'"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 404 {object} models.ErrorResponse "Подключение не найдено"
// @Router /connect/check [post]
func (h *Handler) CheckConnection(c *gin.Context) {
	var req models.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, err, "Missing or invalid SessionID")
		return
	}

	connInfo, err := h.usecase.CheckConnection(req.SessionID)

	if connInfo == nil {
		h.NotFound(c, err)
		return
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "unhealthy", "error": err.Error(), "connection_info": connInfo})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy", "connection_info": connInfo})
}
