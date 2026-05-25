package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mundoinvest/cliente/application"
	"github.com/mundoinvest/shared"
)

type Handler struct {
	criarCmd *application.CriarClienteHandler
}

func NewHandler(criarCmd *application.CriarClienteHandler) *Handler {
	return &Handler{criarCmd: criarCmd}
}

func (h *Handler) Criar(c *gin.Context) {
	var input application.CriarClienteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, shared.ValidationError())
		return
	}

	cliente, err := h.criarCmd.Handle(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.InternalError())
		return
	}

	c.JSON(http.StatusCreated, shared.Success(cliente))
}
