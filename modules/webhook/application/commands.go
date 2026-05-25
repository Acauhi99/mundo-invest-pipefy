package application

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mundoinvest/cliente/domain"
	"github.com/mundoinvest/pipefy"
	webhookDomain "github.com/mundoinvest/webhook/domain"
)

const (
	PriorityAlta     = "prioridade_alta"
	PriorityNormal   = "prioridade_normal"
	StatusProcessado = "Processado"
)

type EventRepository interface {
	IsEventProcessed(eventID string) (bool, error)
	MarkEventProcessed(eventID string) error
	Migrate() error
}

type ClienteQuerier interface {
	Handle(email string) (*domain.Cliente, error)
}

type ClienteUpdater interface {
	UpdateStatusAndPriority(email, status, prioridade string) error
}

type ProcessarCardUpdatedHandler struct {
	eventRepo  EventRepository
	clienteQry ClienteQuerier
	clienteUpd ClienteUpdater
	pipefy     pipefy.PipefyClient
}

func NewProcessarCardUpdatedHandler(
	eventRepo EventRepository,
	clienteQry ClienteQuerier,
	clienteUpd ClienteUpdater,
	pc pipefy.PipefyClient,
) *ProcessarCardUpdatedHandler {
	return &ProcessarCardUpdatedHandler{
		eventRepo:  eventRepo,
		clienteQry: clienteQry,
		clienteUpd: clienteUpd,
		pipefy:     pc,
	}
}

func (h *ProcessarCardUpdatedHandler) Handle(input webhookDomain.CardUpdatedInput) error {
	alreadyProcessed, err := h.eventRepo.IsEventProcessed(input.EventID)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if alreadyProcessed {
		return fmt.Errorf("event %s: %w", input.EventID, webhookDomain.ErrEventAlreadyProcessed)
	}

	c, err := h.clienteQry.Handle(input.ClienteEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("client not found for email %s: %w", input.ClienteEmail, domain.ErrClientNotFound)
		}
		return fmt.Errorf("failed to find client: %w", err)
	}

	prioridade := PriorityNormal
	if c.ValorPatrimonio >= 200000 {
		prioridade = PriorityAlta
	}

	if err := h.clienteUpd.UpdateStatusAndPriority(c.Email, StatusProcessado, prioridade); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	pipefyPayload := h.buildUpdateCardFieldPayload(input.CardID, prioridade)
	h.pipefy.SimulateSend(pipefyPayload)

	if err := h.eventRepo.MarkEventProcessed(input.EventID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}

func (h *ProcessarCardUpdatedHandler) buildUpdateCardFieldPayload(cardID, prioridade string) map[string]interface{} {
	payloads := []map[string]interface{}{
		h.pipefy.BuildUpdateCardFieldPayload(pipefy.UpdateCardFieldInput{
			CardID:   cardID,
			FieldID:  "status",
			NewValue: StatusProcessado,
		}),
		h.pipefy.BuildUpdateCardFieldPayload(pipefy.UpdateCardFieldInput{
			CardID:   cardID,
			FieldID:  "prioridade",
			NewValue: prioridade,
		}),
	}
	return map[string]interface{}{
		"mutations": payloads,
	}
}
