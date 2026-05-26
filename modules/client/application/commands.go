package application

import (
	"fmt"
	"time"

	"github.com/mundoinvest/client/domain"
	"github.com/mundoinvest/pipefy"
)

type CreateClientInput struct {
	Name        string  `json:"cliente_nome" binding:"required"`
	Email       string  `json:"cliente_email" binding:"required,email"`
	RequestType string  `json:"tipo_solicitacao" binding:"required"`
	NetWorth    float64 `json:"valor_patrimonio" binding:"required,gt=0"`
}

type Repository interface {
	Create(c *domain.Client) error
	FindByEmail(email string) (*domain.Client, error)
	UpdateStatusAndPriority(email, status, priority string) error
	UpdateCardID(email, cardID string) error
	Migrate() error
}

type CreateClientHandler struct {
	repo   Repository
	pipefy pipefy.PipefyClient
}

func NewCreateClientHandler(repo Repository, pc pipefy.PipefyClient) *CreateClientHandler {
	return &CreateClientHandler{repo: repo, pipefy: pc}
}

func (h *CreateClientHandler) Handle(input CreateClientInput) (*domain.Client, error) {
	c := &domain.Client{
		Name:        input.Name,
		Email:       input.Email,
		RequestType: input.RequestType,
		NetWorth:    input.NetWorth,
		Status:      "Aguardando Análise",
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

	// TODO: publish to event bus (SQS/SNS)
	_ = domain.ClientCreated{
		ClientID:  c.ID,
		Email:     c.Email,
		Timestamp: time.Now(),
	}

	return c, nil
}

func (h *CreateClientHandler) buildCreateCardPayload(c *domain.Client) map[string]interface{} {
	return h.pipefy.BuildCreateCardPayload(pipefy.CreateCardInput{
		PipeID: 123,
		Title:  c.Name,
		FieldsAttributes: []pipefy.FieldAttribute{
			{FieldID: "nome", FieldValue: c.Name},
			{FieldID: "email", FieldValue: c.Email},
			{FieldID: "tipo_solicitacao", FieldValue: c.RequestType},
			{FieldID: "valor_patrimonio", FieldValue: fmt.Sprintf("%.2f", c.NetWorth)},
		},
	})
}
