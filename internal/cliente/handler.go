package cliente

import (
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

func (h *Handler) Criar(c *gin.Context) {
	var input CriarClienteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ValidationError())
		return
	}

	cliente, err := h.svc.Criar(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError())
		return
	}

	c.JSON(http.StatusCreated, response.Success(cliente))
}
