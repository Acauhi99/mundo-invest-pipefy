package webhook

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mundoinvest/client-management/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CardUpdated(c *gin.Context) {
	var input CardUpdatedInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ValidationError())
		return
	}

	if err := h.svc.ProcessarCardUpdated(input); err != nil {
		switch {
		case errors.Is(err, ErrEventAlreadyProcessed):
			c.JSON(http.StatusConflict, response.ConflictError("EVENT_ALREADY_PROCESSED", "event already processed"))
		case errors.Is(err, ErrClientNotFound):
			c.JSON(http.StatusNotFound, response.NotFoundError("CLIENT_NOT_FOUND", "client not found"))
		default:
			c.JSON(http.StatusInternalServerError, response.InternalError())
		}
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "event processed successfully"}))
}
