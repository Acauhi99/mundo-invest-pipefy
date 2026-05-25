package webhook

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mundoinvest/client-management/internal/cliente"
	"github.com/mundoinvest/client-management/internal/pipefy"
)

const (
	PriorityAlta     = "prioridade_alta"
	PriorityNormal   = "prioridade_normal"
	StatusProcessado = "Processado"
)

var (
	ErrEventAlreadyProcessed = errors.New("event already processed")
	ErrClientNotFound        = errors.New("client not found")
)

type eventTracker interface {
	IsEventProcessed(eventID string) (bool, error)
	MarkEventProcessed(eventID string) error
}

type clientUpdater interface {
	FindByEmail(email string) (*cliente.Cliente, error)
	UpdateStatusAndPriority(email, status, prioridade string) error
}

type pipefyClient interface {
	SimulateSend(payload map[string]interface{}) string
	BuildUpdateCardFieldPayload(input pipefy.UpdateCardFieldInput) map[string]interface{}
}

type Service struct {
	eventRepo    eventTracker
	clienteRepo  clientUpdater
	pipefyClient pipefyClient
}

func NewService(eventRepo eventTracker, clienteRepo clientUpdater, pc pipefyClient) *Service {
	return &Service{
		eventRepo:    eventRepo,
		clienteRepo:  clienteRepo,
		pipefyClient: pc,
	}
}

func (s *Service) ProcessarCardUpdated(input CardUpdatedInput) error {
	alreadyProcessed, err := s.eventRepo.IsEventProcessed(input.EventID)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if alreadyProcessed {
		return fmt.Errorf("event %s: %w", input.EventID, ErrEventAlreadyProcessed)
	}

	c, err := s.clienteRepo.FindByEmail(input.ClienteEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("client not found for email %s: %w", input.ClienteEmail, ErrClientNotFound)
		}
		return fmt.Errorf("failed to find client: %w", err)
	}

	prioridade := PriorityNormal
	if c.ValorPatrimonio >= 200000 {
		prioridade = PriorityAlta
	}

	if err := s.clienteRepo.UpdateStatusAndPriority(c.Email, StatusProcessado, prioridade); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	pipefyPayload := s.buildUpdateCardFieldPayload(input.CardID, prioridade)
	s.pipefyClient.SimulateSend(pipefyPayload)

	if err := s.eventRepo.MarkEventProcessed(input.EventID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}

func (s *Service) buildUpdateCardFieldPayload(cardID, prioridade string) map[string]interface{} {
	payloads := []map[string]interface{}{
		s.pipefyClient.BuildUpdateCardFieldPayload(pipefy.UpdateCardFieldInput{
			CardID:   cardID,
			FieldID:  "status",
			NewValue: StatusProcessado,
		}),
		s.pipefyClient.BuildUpdateCardFieldPayload(pipefy.UpdateCardFieldInput{
			CardID:   cardID,
			FieldID:  "prioridade",
			NewValue: prioridade,
		}),
	}
	return map[string]interface{}{
		"mutations": payloads,
	}
}
