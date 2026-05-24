package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	if err := h.svc.ProcessarCardUpdated(input); err != nil {
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "evento processado com sucesso"})
}
