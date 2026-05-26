package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mundoinvest/client/application"
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
		c.JSON(http.StatusBadRequest, shared.ValidationError())
		return
	}

	client, err := h.createCmd.Handle(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.InternalError())
		return
	}

	c.JSON(http.StatusCreated, shared.Success(client))
}
