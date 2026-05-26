package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	clientDomain "github.com/mundoinvest/client/domain"
	"github.com/mundoinvest/shared"
	"github.com/mundoinvest/webhook/application"
	"github.com/mundoinvest/webhook/domain"
)

type Handler struct {
	processCmd *application.ProcessCardUpdatedHandler
}

func NewHandler(processCmd *application.ProcessCardUpdatedHandler) *Handler {
	return &Handler{processCmd: processCmd}
}

func (h *Handler) CardUpdated(c *gin.Context) {
	var input domain.CardUpdatedInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, shared.ValidationError(err.Error()))
		return
	}

	if err := h.processCmd.Handle(input); err != nil {
		switch {
		case errors.Is(err, domain.ErrEventAlreadyProcessed):
			c.JSON(http.StatusConflict, shared.ConflictError("EVENT_ALREADY_PROCESSED", "event already processed"))
		case errors.Is(err, clientDomain.ErrClientNotFound):
			c.JSON(http.StatusNotFound, shared.NotFoundError("CLIENT_NOT_FOUND", "client not found"))
		default:
			c.JSON(http.StatusInternalServerError, shared.InternalError("failed to process event"))
		}
		return
	}

	c.JSON(http.StatusOK, shared.Success(gin.H{"message": "event processed successfully"}))
}
