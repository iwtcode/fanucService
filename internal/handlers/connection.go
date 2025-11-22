package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

type ConnectionHandler struct {
	usecase interfaces.ConnectionUsecase
}

func NewConnectionHandler(usecase interfaces.ConnectionUsecase) *ConnectionHandler {
	return &ConnectionHandler{usecase: usecase}
}

// Create
// @Summary Create a new connection
// @Description Connects to a Fanuc machine. Throws 500 if connection takes longer than 5 seconds.
// @Tags Connection
// @Accept json
// @Produce json
// @Param input body fanucService.ConnectionRequest true "Connection Data"
// @Security ApiKeyAuth
// @Success 200 {object} fanucService.ConnectionResponse
// @Failure 400 {object} fanucService.ConnectionResponse
// @Failure 500 {object} fanucService.ConnectionResponse
// @Router /api/v1/connect [post]
func (h *ConnectionHandler) Create(c *gin.Context) {
	var req fanucService.ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, fanucService.ConnectionResponse{Status: "error", Message: err.Error()})
		return
	}

	machine, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, fanucService.ConnectionResponse{Status: "error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, fanucService.ConnectionResponse{Status: "ok", Data: machine})
}

// Get
// @Summary Get connections or Check specific connection
// @Description If 'id' is provided, checks health of specific connection. If not, lists all connections.
// @Tags Connection
// @Produce json
// @Param id query string false "Machine ID (optional)"
// @Security ApiKeyAuth
// @Success 200 {object} fanucService.ConnectionResponse
// @Router /api/v1/connect [get]
func (h *ConnectionHandler) Get(c *gin.Context) {
	id := c.Query("id")

	// Если ID передан, выполняем проверку конкретного подключения
	if id != "" {
		machine, err := h.usecase.Check(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, fanucService.ConnectionResponse{
				Status:  "error",
				Message: err.Error(),
				Data:    machine,
			})
			return
		}
		c.JSON(http.StatusOK, fanucService.ConnectionResponse{Status: "ok", Data: machine})
		return
	}

	// Иначе возвращаем список всех подключений
	machines, err := h.usecase.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, fanucService.ConnectionResponse{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, fanucService.ConnectionResponse{Status: "ok", Data: machines})
}

// Delete
// @Summary Delete connection
// @Tags Connection
// @Param id query string true "Machine ID"
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} fanucService.ConnectionResponse
// @Router /api/v1/connect [delete]
func (h *ConnectionHandler) Delete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, fanucService.ConnectionResponse{Status: "error", Message: "id is required"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, fanucService.ConnectionResponse{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, fanucService.ConnectionResponse{Status: "ok", Message: "deleted"})
}
