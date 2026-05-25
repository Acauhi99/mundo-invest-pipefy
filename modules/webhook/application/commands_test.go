package application_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mundoinvest/cliente/domain"
	"github.com/mundoinvest/pipefy"
	webhookApp "github.com/mundoinvest/webhook/application"
	webhookDomain "github.com/mundoinvest/webhook/domain"
)

type mockEventRepository struct {
	isProcessedFn   func(eventID string) (bool, error)
	markProcessedFn func(eventID string) error
}

func (m *mockEventRepository) IsEventProcessed(eventID string) (bool, error) {
	return m.isProcessedFn(eventID)
}
func (m *mockEventRepository) MarkEventProcessed(eventID string) error {
	return m.markProcessedFn(eventID)
}
func (m *mockEventRepository) Migrate() error { return nil }

type mockClienteQuerier struct {
	handleFn func(email string) (*domain.Cliente, error)
}

func (m *mockClienteQuerier) Handle(email string) (*domain.Cliente, error) {
	return m.handleFn(email)
}

type mockClienteUpdater struct {
	updateFn func(email, status, prioridade string) error
}

func (m *mockClienteUpdater) UpdateStatusAndPriority(email, status, prioridade string) error {
	return m.updateFn(email, status, prioridade)
}

type mockPipefy struct {
	simulateSendFn                func(payload map[string]interface{}) string
	buildUpdateCardFieldPayloadFn func(input pipefy.UpdateCardFieldInput) map[string]interface{}
}

func (m *mockPipefy) SimulateSend(payload map[string]interface{}) string {
	return m.simulateSendFn(payload)
}
func (m *mockPipefy) BuildCreateCardPayload(input pipefy.CreateCardInput) map[string]interface{} {
	return map[string]interface{}{}
}
func (m *mockPipefy) BuildUpdateCardFieldPayload(input pipefy.UpdateCardFieldInput) map[string]interface{} {
	return m.buildUpdateCardFieldPayloadFn(input)
}

func TestProcessar_HighPriority(t *testing.T) {
	var updatedStatus, updatedPriority string
	clienteUpdater := &mockClienteUpdater{
		updateFn: func(email, status, prioridade string) error {
			updatedStatus = status
			updatedPriority = prioridade
			return nil
		},
	}
	clienteQuerier := &mockClienteQuerier{
		handleFn: func(email string) (*domain.Cliente, error) {
			return &domain.Cliente{
				Email:           "joao@example.com",
				ValorPatrimonio: 250000,
			}, nil
		},
	}
	eventRepo := &mockEventRepository{
		isProcessedFn:   func(eventID string) (bool, error) { return false, nil },
		markProcessedFn: func(eventID string) error { return nil },
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "" },
		buildUpdateCardFieldPayloadFn: func(input pipefy.UpdateCardFieldInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	handler := webhookApp.NewProcessarCardUpdatedHandler(eventRepo, clienteQuerier, clienteUpdater, pipefyMock)
	err := handler.Handle(webhookDomain.CardUpdatedInput{
		EventID:      "evt_001",
		CardID:       "card_001",
		ClienteEmail: "joao@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedStatus != "Processado" {
		t.Errorf("expected status 'Processado', got '%s'", updatedStatus)
	}
	if updatedPriority != "prioridade_alta" {
		t.Errorf("expected priority 'prioridade_alta', got '%s'", updatedPriority)
	}
}

func TestProcessar_NormalPriority(t *testing.T) {
	var updatedPriority string
	clienteUpdater := &mockClienteUpdater{
		updateFn: func(email, status, prioridade string) error {
			updatedPriority = prioridade
			return nil
		},
	}
	clienteQuerier := &mockClienteQuerier{
		handleFn: func(email string) (*domain.Cliente, error) {
			return &domain.Cliente{
				Email:           "maria@example.com",
				ValorPatrimonio: 50000,
			}, nil
		},
	}
	eventRepo := &mockEventRepository{
		isProcessedFn:   func(eventID string) (bool, error) { return false, nil },
		markProcessedFn: func(eventID string) error { return nil },
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "" },
		buildUpdateCardFieldPayloadFn: func(input pipefy.UpdateCardFieldInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	handler := webhookApp.NewProcessarCardUpdatedHandler(eventRepo, clienteQuerier, clienteUpdater, pipefyMock)
	err := handler.Handle(webhookDomain.CardUpdatedInput{
		EventID:      "evt_002",
		CardID:       "card_002",
		ClienteEmail: "maria@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedPriority != "prioridade_normal" {
		t.Errorf("expected priority 'prioridade_normal', got '%s'", updatedPriority)
	}
}

func TestProcessar_AlreadyProcessed(t *testing.T) {
	eventRepo := &mockEventRepository{
		isProcessedFn: func(eventID string) (bool, error) { return true, nil },
	}
	clienteQuerier := &mockClienteQuerier{}
	clienteUpdater := &mockClienteUpdater{}
	pipefyMock := &mockPipefy{}

	handler := webhookApp.NewProcessarCardUpdatedHandler(eventRepo, clienteQuerier, clienteUpdater, pipefyMock)
	err := handler.Handle(webhookDomain.CardUpdatedInput{
		EventID:      "evt_dup",
		CardID:       "card_001",
		ClienteEmail: "joao@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err == nil {
		t.Fatal("expected error for duplicate event")
	}
	if !errors.Is(err, webhookDomain.ErrEventAlreadyProcessed) {
		t.Errorf("expected ErrEventAlreadyProcessed, got %v", err)
	}
}

func TestProcessar_ClientNotFound(t *testing.T) {
	eventRepo := &mockEventRepository{
		isProcessedFn: func(eventID string) (bool, error) { return false, nil },
	}
	clienteQuerier := &mockClienteQuerier{
		handleFn: func(email string) (*domain.Cliente, error) {
			return nil, sql.ErrNoRows
		},
	}
	clienteUpdater := &mockClienteUpdater{}
	pipefyMock := &mockPipefy{}

	handler := webhookApp.NewProcessarCardUpdatedHandler(eventRepo, clienteQuerier, clienteUpdater, pipefyMock)
	err := handler.Handle(webhookDomain.CardUpdatedInput{
		EventID:      "evt_004",
		CardID:       "card_001",
		ClienteEmail: "ghost@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err == nil {
		t.Fatal("expected error for missing client")
	}
	if !errors.Is(err, domain.ErrClientNotFound) {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
}

func TestProcessar_UpdateError(t *testing.T) {
	eventRepo := &mockEventRepository{
		isProcessedFn:   func(eventID string) (bool, error) { return false, nil },
		markProcessedFn: func(eventID string) error { return nil },
	}
	clienteQuerier := &mockClienteQuerier{
		handleFn: func(email string) (*domain.Cliente, error) {
			return &domain.Cliente{Email: email, ValorPatrimonio: 300000}, nil
		},
	}
	clienteUpdater := &mockClienteUpdater{
		updateFn: func(email, status, prioridade string) error {
			return errors.New("db write error")
		},
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "" },
		buildUpdateCardFieldPayloadFn: func(input pipefy.UpdateCardFieldInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	handler := webhookApp.NewProcessarCardUpdatedHandler(eventRepo, clienteQuerier, clienteUpdater, pipefyMock)
	err := handler.Handle(webhookDomain.CardUpdatedInput{
		EventID:      "evt_005",
		CardID:       "card_001",
		ClienteEmail: "joao@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err == nil {
		t.Fatal("expected error on update failure")
	}
	if !strings.Contains(err.Error(), "failed to update client") {
		t.Errorf("expected 'failed to update client' in error, got: %v", err)
	}
}
