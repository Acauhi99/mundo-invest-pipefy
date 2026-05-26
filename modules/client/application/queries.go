package application

import (
	"github.com/mundoinvest/client/domain"
)

type GetClientByEmailHandler struct {
	repo Repository
}

func NewGetClientByEmailHandler(repo Repository) *GetClientByEmailHandler {
	return &GetClientByEmailHandler{repo: repo}
}

func (h *GetClientByEmailHandler) Handle(email string) (*domain.Client, error) {
	c, err := h.repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	return c, nil
}
