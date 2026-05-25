package cliente

import (
	"errors"
	"testing"
	"time"

	"github.com/mundoinvest/client-management/internal/pipefy"
)

type mockPersister struct {
	createFn       func(c *Cliente) error
	updateCardIDFn func(email, cardID string) error
}

func (m *mockPersister) Create(c *Cliente) error { return m.createFn(c) }
func (m *mockPersister) UpdateCardID(email, cardID string) error {
	return m.updateCardIDFn(email, cardID)
}

type mockPipefy struct {
	simulateSendFn           func(payload map[string]interface{}) string
	buildCreateCardPayloadFn func(input pipefy.CreateCardInput) map[string]interface{}
}

func (m *mockPipefy) SimulateSend(payload map[string]interface{}) string {
	return m.simulateSendFn(payload)
}
func (m *mockPipefy) BuildCreateCardPayload(input pipefy.CreateCardInput) map[string]interface{} {
	return m.buildCreateCardPayloadFn(input)
}

func TestCriar_Success(t *testing.T) {
	persister := &mockPersister{
		createFn: func(c *Cliente) error {
			c.ID = 1
			c.CreatedAt = time.Now()
			return nil
		},
		updateCardIDFn: func(email, cardID string) error { return nil },
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "card_sim_test_1" },
		buildCreateCardPayloadFn: func(input pipefy.CreateCardInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	svc := NewService(persister, pipefyMock)
	input := CriarClienteInput{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: 250000,
	}

	c, err := svc.Criar(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.ID != 1 {
		t.Errorf("expected ID 1, got %d", c.ID)
	}
	if c.Status != "Aguardando Análise" {
		t.Errorf("expected status 'Aguardando Análise', got '%s'", c.Status)
	}
	if c.CardID != "card_sim_test_1" {
		t.Errorf("expected cardID 'card_sim_test_1', got '%s'", c.CardID)
	}
}

func TestCriar_RepoError(t *testing.T) {
	persister := &mockPersister{
		createFn: func(c *Cliente) error {
			return errors.New("db error")
		},
		updateCardIDFn: func(email, cardID string) error { return nil },
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "card_sim_test_1" },
		buildCreateCardPayloadFn: func(input pipefy.CreateCardInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	svc := NewService(persister, pipefyMock)
	input := CriarClienteInput{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: 250000,
	}

	_, err := svc.Criar(input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCriar_UpdateCardIDError(t *testing.T) {
	persister := &mockPersister{
		createFn: func(c *Cliente) error {
			c.ID = 1
			c.CreatedAt = time.Now()
			return nil
		},
		updateCardIDFn: func(email, cardID string) error {
			return errors.New("update card error")
		},
	}
	pipefyMock := &mockPipefy{
		simulateSendFn: func(payload map[string]interface{}) string { return "card_sim_test_1" },
		buildCreateCardPayloadFn: func(input pipefy.CreateCardInput) map[string]interface{} {
			return map[string]interface{}{}
		},
	}

	svc := NewService(persister, pipefyMock)
	input := CriarClienteInput{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: 250000,
	}

	_, err := svc.Criar(input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
