package application

import (
	"errors"
	"fmt"
	"time"

	"github.com/mundoinvest/client/domain"
	"github.com/mundoinvest/pipefy"
	webhookDomain "github.com/mundoinvest/webhook/domain"
)

const (
	PriorityHigh    = "prioridade_alta"
	PriorityNormal  = "prioridade_normal"
	StatusProcessed = "Processado"
)

type EventRepository interface {
	IsEventProcessed(eventID string) (bool, error)
	MarkEventProcessed(eventID string) error
	Migrate() error
}

type ClientQuerier interface {
	Handle(email string) (*domain.Client, error)
}

type ClientUpdater interface {
	UpdateStatusAndPriority(email, status, priority string) error
}

type ProcessCardUpdatedHandler struct {
	eventRepo EventRepository
	clientQry ClientQuerier
	clientUpd ClientUpdater
	pipefy    pipefy.PipefyClient
}

func NewProcessCardUpdatedHandler(
	eventRepo EventRepository,
	clientQry ClientQuerier,
	clientUpd ClientUpdater,
	pc pipefy.PipefyClient,
) *ProcessCardUpdatedHandler {
	return &ProcessCardUpdatedHandler{
		eventRepo: eventRepo,
		clientQry: clientQry,
		clientUpd: clientUpd,
		pipefy:    pc,
	}
}

func (h *ProcessCardUpdatedHandler) Handle(input webhookDomain.CardUpdatedInput) error {
	alreadyProcessed, err := h.eventRepo.IsEventProcessed(input.EventID)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if alreadyProcessed {
		return fmt.Errorf("event %s: %w", input.EventID, webhookDomain.ErrEventAlreadyProcessed)
	}

	c, err := h.clientQry.Handle(input.ClienteEmail)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			return fmt.Errorf("client not found for email %s: %w", input.ClienteEmail, domain.ErrClientNotFound)
		}
		return fmt.Errorf("failed to find client: %w", err)
	}

	priority := PriorityNormal
	if c.NetWorth >= 200000 {
		priority = PriorityHigh
	}

	if err := h.clientUpd.UpdateStatusAndPriority(c.Email, StatusProcessed, priority); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	pipefyPayload := h.buildUpdateCardFieldPayload(input.CardID, priority)
	h.pipefy.SimulateSend(pipefyPayload)

	if err := h.eventRepo.MarkEventProcessed(input.EventID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	// TODO: publish to event bus (SQS/SNS)
	_ = domain.ClientProcessed{
		Email:     input.ClienteEmail,
		Priority:  priority,
		Timestamp: time.Now(),
	}

	return nil
}

func (h *ProcessCardUpdatedHandler) buildUpdateCardFieldPayload(cardID, priority string) map[string]interface{} {
	payloads := []map[string]interface{}{
		h.pipefy.BuildUpdateCardFieldPayload(pipefy.UpdateCardFieldInput{
			CardID:   cardID,
			FieldID:  "status",
			NewValue: StatusProcessed,
		}),
		h.pipefy.BuildUpdateCardFieldPayload(pipefy.UpdateCardFieldInput{
			CardID:   cardID,
			FieldID:  "prioridade",
			NewValue: priority,
		}),
	}
	return map[string]interface{}{
		"mutations": payloads,
	}
}
