package webhook

import (
	"fmt"

	"github.com/mundoinvest/client-management/internal/cliente"
	"github.com/mundoinvest/client-management/internal/pipefy"
)

const (
	PriorityAlta     = "prioridade_alta"
	PriorityNormal   = "prioridade_normal"
	StatusProcessado = "Processado"
)

type Service struct {
	eventRepo    *Repository
	clienteRepo  *cliente.Repository
	pipefyClient *pipefy.Client
}

func NewService(eventRepo *Repository, clienteRepo *cliente.Repository, pipefyClient *pipefy.Client) *Service {
	return &Service{
		eventRepo:    eventRepo,
		clienteRepo:  clienteRepo,
		pipefyClient: pipefyClient,
	}
}

func (s *Service) ProcessarCardUpdated(input CardUpdatedInput) error {
	alreadyProcessed, err := s.eventRepo.IsEventProcessed(input.EventID)
	if err != nil {
		return fmt.Errorf("erro ao verificar idempotência: %w", err)
	}
	if alreadyProcessed {
		return fmt.Errorf("evento %s já processado", input.EventID)
	}

	c, err := s.clienteRepo.FindByEmail(input.ClienteEmail)
	if err != nil {
		return fmt.Errorf("cliente não encontrado para o email %s: %w", input.ClienteEmail, err)
	}

	prioridade := PriorityNormal
	if c.ValorPatrimonio >= 200000 {
		prioridade = PriorityAlta
	}

	if err := s.clienteRepo.UpdateStatusAndPriority(c.Email, StatusProcessado, prioridade); err != nil {
		return fmt.Errorf("erro ao atualizar cliente: %w", err)
	}

	pipefyPayload := s.buildUpdateCardFieldPayload(input.CardID, prioridade)
	s.pipefyClient.SimulateSend(pipefyPayload)

	if err := s.eventRepo.MarkEventProcessed(input.EventID); err != nil {
		return fmt.Errorf("erro ao marcar evento como processado: %w", err)
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
