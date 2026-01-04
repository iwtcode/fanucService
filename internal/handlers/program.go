package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

type ProgramHandler struct {
	usecase interfaces.ProgramUsecase
}

func NewProgramHandler(usecase interfaces.ProgramUsecase) *ProgramHandler {
	return &ProgramHandler{usecase: usecase}
}

// Get
// @Summary Get full control program
// @Description Returns the raw text of the current executing program (G-Code)
// @Tags Program
// @Produce plain
// @Param id query string true "Machine ID"
// @Security ApiKeyAuth
// @Success 200 {string} string "Program content"
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/program [get]
func (h *ProgramHandler) Get(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, "id is required")
		return
	}

	program, err := h.usecase.GetProgram(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, program)
}
