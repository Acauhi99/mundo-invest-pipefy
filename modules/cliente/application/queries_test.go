package application_test

import (
	"database/sql"
	"testing"

	clienteApp "github.com/mundoinvest/cliente/application"
	"github.com/mundoinvest/cliente/domain"
)

func TestObterClientePorEmail_Success(t *testing.T) {
	t.Skip("requires integration test with real database")
}

func TestObterClientePorEmail_NotFound(t *testing.T) {
	t.Skip("requires integration test with real database")
}

type mockQueryRepo struct {
	findFn func(email string) (*domain.Cliente, error)
}

func (m *mockQueryRepo) Create(c *domain.Cliente) error                                 { return nil }
func (m *mockQueryRepo) FindByEmail(email string) (*domain.Cliente, error)              { return m.findFn(email) }
func (m *mockQueryRepo) UpdateStatusAndPriority(email, status, prioridade string) error { return nil }
func (m *mockQueryRepo) UpdateCardID(email, cardID string) error                        { return nil }
func (m *mockQueryRepo) Migrate() error                                                 { return nil }

func TestObterCliente_ReturnsError(t *testing.T) {
	repo := &mockQueryRepo{
		findFn: func(email string) (*domain.Cliente, error) {
			return nil, sql.ErrNoRows
		},
	}
	handler := clienteApp.NewObterClientePorEmailHandler(repo)
	_, err := handler.Handle("missing@example.com")
	if err == nil {
		t.Fatal("expected error for missing client")
	}
}
