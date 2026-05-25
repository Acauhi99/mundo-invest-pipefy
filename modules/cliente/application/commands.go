package application

import (
	"fmt"

	"github.com/mundoinvest/cliente/domain"
	"github.com/mundoinvest/pipefy"
)

type CriarClienteInput struct {
	Nome            string  `json:"cliente_nome" binding:"required"`
	Email           string  `json:"cliente_email" binding:"required,email"`
	TipoSolicitacao string  `json:"tipo_solicitacao" binding:"required"`
	ValorPatrimonio float64 `json:"valor_patrimonio" binding:"required,gt=0"`
}

type Repository interface {
	Create(c *domain.Cliente) error
	FindByEmail(email string) (*domain.Cliente, error)
	UpdateStatusAndPriority(email, status, prioridade string) error
	UpdateCardID(email, cardID string) error
	Migrate() error
}

type CriarClienteHandler struct {
	repo   Repository
	pipefy pipefy.PipefyClient
}

func NewCriarClienteHandler(repo Repository, pc pipefy.PipefyClient) *CriarClienteHandler {
	return &CriarClienteHandler{repo: repo, pipefy: pc}
}

func (h *CriarClienteHandler) Handle(input CriarClienteInput) (*domain.Cliente, error) {
	c := &domain.Cliente{
		Nome:            input.Nome,
		Email:           input.Email,
		TipoSolicitacao: input.TipoSolicitacao,
		ValorPatrimonio: input.ValorPatrimonio,
		Status:          "Aguardando Análise",
	}

	if err := h.repo.Create(c); err != nil {
		return nil, fmt.Errorf("failed to persist client: %w", err)
	}

	pipefyPayload := h.buildCreateCardPayload(c)
	cardID := h.pipefy.SimulateSend(pipefyPayload)

	if err := h.repo.UpdateCardID(c.Email, cardID); err != nil {
		return nil, fmt.Errorf("failed to update card_id: %w", err)
	}
	c.CardID = cardID

	return c, nil
}

func (h *CriarClienteHandler) buildCreateCardPayload(c *domain.Cliente) map[string]interface{} {
	return h.pipefy.BuildCreateCardPayload(pipefy.CreateCardInput{
		PipeID: 123,
		Title:  c.Nome,
		FieldsAttributes: []pipefy.FieldAttribute{
			{FieldID: "nome", FieldValue: c.Nome},
			{FieldID: "email", FieldValue: c.Email},
			{FieldID: "tipo_solicitacao", FieldValue: c.TipoSolicitacao},
			{FieldID: "valor_patrimonio", FieldValue: fmt.Sprintf("%.2f", c.ValorPatrimonio)},
		},
	})
}
