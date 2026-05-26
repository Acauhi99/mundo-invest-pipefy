package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mundoinvest/client/application"
	"github.com/mundoinvest/client/domain"
	"github.com/mundoinvest/shared"
)

type Handler struct {
	createCmd *application.CreateClientHandler
}

func NewHandler(createCmd *application.CreateClientHandler) *Handler {
	return &Handler{createCmd: createCmd}
}

func (h *Handler) Create(c *gin.Context) {
	var input application.CreateClientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, shared.ValidationError(err.Error()))
		return
	}

	client, err := h.createCmd.Handle(input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, shared.ConflictError("EMAIL_ALREADY_EXISTS", err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, shared.InternalError("failed to create client"))
		}
		return
	}

	c.JSON(http.StatusCreated, shared.Success(client))
}
