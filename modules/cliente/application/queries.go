package application

import (
	"fmt"

	"github.com/mundoinvest/cliente/domain"
)

type ObterClientePorEmailHandler struct {
	repo Repository
}

func NewObterClientePorEmailHandler(repo Repository) *ObterClientePorEmailHandler {
	return &ObterClientePorEmailHandler{repo: repo}
}

func (h *ObterClientePorEmailHandler) Handle(email string) (*domain.Cliente, error) {
	c, err := h.repo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}
	return c, nil
}
