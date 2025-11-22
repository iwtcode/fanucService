package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService/internal/domain/models"
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
// @Description Connects to a Fanuc machine
// @Tags Connection
// @Accept json
// @Produce json
// @Param input body models.ConnectionRequest true "Connection Data"
// @Security ApiKeyAuth
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/connect [post]
func (h *ConnectionHandler) Create(c *gin.Context) {
	var req models.ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	machine, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(c, machine)
}

// Get
// @Summary Get connections or Check specific connection
// @Description If 'id' is provided, checks health of specific connection. If not, lists all connections.
// @Tags Connection
// @Produce json
// @Param id query string false "Machine ID (optional)"
// @Security ApiKeyAuth
// @Success 200 {object} models.APIResponse
// @Router /api/v1/connect [get]
func (h *ConnectionHandler) Get(c *gin.Context) {
	id := c.Query("id")

	// Check specific connection
	if id != "" {
		machine, err := h.usecase.Check(c.Request.Context(), id)
		if err != nil {
			RespondError(c, http.StatusServiceUnavailable, err.Error(), machine)
			return
		}
		RespondSuccess(c, machine)
		return
	}

	// List all
	machines, err := h.usecase.List(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	RespondSuccess(c, machines)
}

// Delete
// @Summary Delete connection
// @Tags Connection
// @Param id query string true "Machine ID"
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.APIResponse
// @Router /api/v1/connect [delete]
func (h *ConnectionHandler) Delete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondMessage(c, fmt.Sprintf("session %s successfully deleted", id))
}
