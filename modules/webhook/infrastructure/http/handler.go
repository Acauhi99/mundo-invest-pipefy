package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	clienteDomain "github.com/mundoinvest/cliente/domain"
	"github.com/mundoinvest/shared"
	"github.com/mundoinvest/webhook/application"
	"github.com/mundoinvest/webhook/domain"
)

type Handler struct {
	processarCmd *application.ProcessarCardUpdatedHandler
}

func NewHandler(processarCmd *application.ProcessarCardUpdatedHandler) *Handler {
	return &Handler{processarCmd: processarCmd}
}

func (h *Handler) CardUpdated(c *gin.Context) {
	var input domain.CardUpdatedInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, shared.ValidationError())
		return
	}

	if err := h.processarCmd.Handle(input); err != nil {
		switch {
		case errors.Is(err, domain.ErrEventAlreadyProcessed):
			c.JSON(http.StatusConflict, shared.ConflictError("EVENT_ALREADY_PROCESSED", "event already processed"))
		case errors.Is(err, clienteDomain.ErrClientNotFound):
			c.JSON(http.StatusNotFound, shared.NotFoundError("CLIENT_NOT_FOUND", "client not found"))
		default:
			c.JSON(http.StatusInternalServerError, shared.InternalError())
		}
		return
	}

	c.JSON(http.StatusOK, shared.Success(gin.H{"message": "event processed successfully"}))
}
