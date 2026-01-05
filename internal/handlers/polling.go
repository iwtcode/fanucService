package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService/internal/domain/models"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

type PollingHandler struct {
	usecase interfaces.PollingUsecase
}

func NewPollingHandler(usecase interfaces.PollingUsecase) *PollingHandler {
	return &PollingHandler{usecase: usecase}
}

// Start
// @Summary Start polling for a machine
// @Description Starts periodic data collection for a specific machine session
// @Tags Polling
// @Accept json
// @Produce json
// @Param input body models.StartPollingRequest true "Polling Config"
// @Security ApiKeyAuth
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/polling/start [post]
func (h *PollingHandler) Start(c *gin.Context) {
	var req models.StartPollingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.Start(c.Request.Context(), req); err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondMessage(c, "Polling started for session "+req.ID)
}

// Stop
// @Summary Stop polling for a machine
// @Description Stops periodic data collection
// @Tags Polling
// @Accept json
// @Produce json
// @Param input body models.StopPollingRequest true "Polling Config"
// @Security ApiKeyAuth
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/polling/stop [post]
func (h *PollingHandler) Stop(c *gin.Context) {
	var req models.StopPollingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.Stop(c.Request.Context(), req); err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondMessage(c, "Polling stopped for session "+req.ID)
}
