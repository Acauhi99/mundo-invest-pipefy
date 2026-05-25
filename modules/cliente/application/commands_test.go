package application_test

import (
	"errors"
	"testing"

	clienteApp "github.com/mundoinvest/cliente/application"
	"github.com/mundoinvest/cliente/domain"
	"github.com/mundoinvest/pipefy"
)

type mockRepository struct {
	createFn       func(c *domain.Cliente) error
	updateCardIDFn func(email, cardID string) error
}

func (m *mockRepository) Create(c *domain.Cliente) error                                 { return m.createFn(c) }
func (m *mockRepository) FindByEmail(email string) (*domain.Cliente, error)              { return nil, nil }
func (m *mockRepository) UpdateStatusAndPriority(email, status, prioridade string) error { return nil }
func (m *mockRepository) UpdateCardID(email, cardID string) error {
	return m.updateCardIDFn(email, cardID)
}
func (m *mockRepository) Migrate() error { return nil }

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
func (m *mockPipefy) BuildUpdateCardFieldPayload(input pipefy.UpdateCardFieldInput) map[string]interface{} {
	return map[string]interface{}{}
}

func TestCriar_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(c *domain.Cliente) error {
			c.ID = 1
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

	handler := clienteApp.NewCriarClienteHandler(repo, pipefyMock)
	input := clienteApp.CriarClienteInput{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: 250000,
	}

	c, err := handler.Handle(input)
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
	repo := &mockRepository{
		createFn: func(c *domain.Cliente) error {
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

	handler := clienteApp.NewCriarClienteHandler(repo, pipefyMock)
	input := clienteApp.CriarClienteInput{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: 250000,
	}

	_, err := handler.Handle(input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCriar_UpdateCardIDError(t *testing.T) {
	repo := &mockRepository{
		createFn: func(c *domain.Cliente) error {
			c.ID = 1
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

	handler := clienteApp.NewCriarClienteHandler(repo, pipefyMock)
	input := clienteApp.CriarClienteInput{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: 250000,
	}

	_, err := handler.Handle(input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
