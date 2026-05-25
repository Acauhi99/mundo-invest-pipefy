package webhook

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mundoinvest/client-management/internal/cliente"
	"github.com/mundoinvest/client-management/internal/pipefy"
)

type mockEventTracker struct {
	isProcessedFn   func(eventID string) (bool, error)
	markProcessedFn func(eventID string) error
}

func (m *mockEventTracker) IsEventProcessed(eventID string) (bool, error) {
	return m.isProcessedFn(eventID)
}
func (m *mockEventTracker) MarkEventProcessed(eventID string) error {
	return m.markProcessedFn(eventID)
}

type mockClientUpdater struct {
	findByEmailFn             func(email string) (*cliente.Cliente, error)
	updateStatusAndPriorityFn func(email, status, prioridade string) error
}

func (m *mockClientUpdater) FindByEmail(email string) (*cliente.Cliente, error) {
	return m.findByEmailFn(email)
}
func (m *mockClientUpdater) UpdateStatusAndPriority(email, status, prioridade string) error {
	return m.updateStatusAndPriorityFn(email, status, prioridade)
}

type mockPipefy struct {
	simulateSendFn                func(payload map[string]interface{}) string
	buildUpdateCardFieldPayloadFn func(input pipefy.UpdateCardFieldInput) map[string]interface{}
}

func (m *mockPipefy) SimulateSend(payload map[string]interface{}) string {
	return m.simulateSendFn(payload)
}
func (m *mockPipefy) BuildUpdateCardFieldPayload(input pipefy.UpdateCardFieldInput) map[string]interface{} {
	return m.buildUpdateCardFieldPayloadFn(input)
}

func TestProcessar_HighPriority(t *testing.T) {
	var updatedStatus, updatedPriority string
	clientUpdater := &mockClientUpdater{
		findByEmailFn: func(email string) (*cliente.Cliente, error) {
			return &cliente.Cliente{
				Email:           "joao@example.com",
				ValorPatrimonio: 250000,
			}, nil
		},
		updateStatusAndPriorityFn: func(email, status, prioridade string) error {
			updatedStatus = status
			updatedPriority = prioridade
			return nil
		},
	}
	eventTracker := &mockEventTracker{
		isProcessedFn:   func(eventID string) (bool, error) { return false, nil },
		markProcessedFn: func(eventID string) error { return nil },
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "" },
		buildUpdateCardFieldPayloadFn: func(input pipefy.UpdateCardFieldInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	svc := NewService(eventTracker, clientUpdater, pipefyMock)
	err := svc.ProcessarCardUpdated(CardUpdatedInput{
		EventID:      "evt_001",
		CardID:       "card_001",
		ClienteEmail: "joao@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedStatus != StatusProcessado {
		t.Errorf("expected status '%s', got '%s'", StatusProcessado, updatedStatus)
	}
	if updatedPriority != PriorityAlta {
		t.Errorf("expected priority '%s', got '%s'", PriorityAlta, updatedPriority)
	}
}

func TestProcessar_NormalPriority(t *testing.T) {
	var updatedPriority string
	clientUpdater := &mockClientUpdater{
		findByEmailFn: func(email string) (*cliente.Cliente, error) {
			return &cliente.Cliente{
				Email:           "maria@example.com",
				ValorPatrimonio: 50000,
			}, nil
		},
		updateStatusAndPriorityFn: func(email, status, prioridade string) error {
			updatedPriority = prioridade
			return nil
		},
	}
	eventTracker := &mockEventTracker{
		isProcessedFn:   func(eventID string) (bool, error) { return false, nil },
		markProcessedFn: func(eventID string) error { return nil },
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "" },
		buildUpdateCardFieldPayloadFn: func(input pipefy.UpdateCardFieldInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	svc := NewService(eventTracker, clientUpdater, pipefyMock)
	err := svc.ProcessarCardUpdated(CardUpdatedInput{
		EventID:      "evt_002",
		CardID:       "card_002",
		ClienteEmail: "maria@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedPriority != PriorityNormal {
		t.Errorf("expected priority '%s', got '%s'", PriorityNormal, updatedPriority)
	}
}

func TestProcessar_AlreadyProcessed(t *testing.T) {
	eventTracker := &mockEventTracker{
		isProcessedFn: func(eventID string) (bool, error) { return true, nil },
	}
	clientUpdater := &mockClientUpdater{}
	pipefyMock := &mockPipefy{}

	svc := NewService(eventTracker, clientUpdater, pipefyMock)
	err := svc.ProcessarCardUpdated(CardUpdatedInput{
		EventID:      "evt_dup",
		CardID:       "card_001",
		ClienteEmail: "joao@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err == nil {
		t.Fatal("expected error for duplicate event")
	}
	if !errors.Is(err, ErrEventAlreadyProcessed) {
		t.Errorf("expected ErrEventAlreadyProcessed, got %v", err)
	}
}

func TestProcessar_ClientNotFound(t *testing.T) {
	eventTracker := &mockEventTracker{
		isProcessedFn: func(eventID string) (bool, error) { return false, nil },
	}
	clientUpdater := &mockClientUpdater{
		findByEmailFn: func(email string) (*cliente.Cliente, error) {
			return nil, sql.ErrNoRows
		},
	}
	pipefyMock := &mockPipefy{}

	svc := NewService(eventTracker, clientUpdater, pipefyMock)
	err := svc.ProcessarCardUpdated(CardUpdatedInput{
		EventID:      "evt_004",
		CardID:       "card_001",
		ClienteEmail: "ghost@example.com",
		Timestamp:    "2026-01-01T00:00:00Z",
	})

	if err == nil {
		t.Fatal("expected error for missing client")
	}
	if !errors.Is(err, ErrClientNotFound) {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
}

func TestProcessar_UpdateError(t *testing.T) {
	eventTracker := &mockEventTracker{
		isProcessedFn:   func(eventID string) (bool, error) { return false, nil },
		markProcessedFn: func(eventID string) error { return nil },
	}
	clientUpdater := &mockClientUpdater{
		findByEmailFn: func(email string) (*cliente.Cliente, error) {
			return &cliente.Cliente{Email: email, ValorPatrimonio: 300000}, nil
		},
		updateStatusAndPriorityFn: func(email, status, prioridade string) error {
			return errors.New("db write error")
		},
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "" },
		buildUpdateCardFieldPayloadFn: func(input pipefy.UpdateCardFieldInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	svc := NewService(eventTracker, clientUpdater, pipefyMock)
	err := svc.ProcessarCardUpdated(CardUpdatedInput{
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
