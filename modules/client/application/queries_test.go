package application_test

import (
	"testing"

	clientApp "github.com/mundoinvest/client/application"
	"github.com/mundoinvest/client/domain"
)

func TestObterClientePorEmail_Success(t *testing.T) {
	t.Skip("requires integration test with real database")
}

func TestObterClientePorEmail_NotFound(t *testing.T) {
	t.Skip("requires integration test with real database")
}

type mockQueryRepo struct {
	findFn func(email string) (*domain.Client, error)
}

func (m *mockQueryRepo) Create(c *domain.Client) error                                { return nil }
func (m *mockQueryRepo) FindByEmail(email string) (*domain.Client, error)             { return m.findFn(email) }
func (m *mockQueryRepo) UpdateStatusAndPriority(email, status, priority string) error { return nil }
func (m *mockQueryRepo) UpdateCardID(email, cardID string) error                      { return nil }
func (m *mockQueryRepo) Migrate() error                                               { return nil }

func TestObterCliente_ReturnsError(t *testing.T) {
	repo := &mockQueryRepo{
		findFn: func(email string) (*domain.Client, error) {
			return nil, domain.ErrClientNotFound
		},
	}
	handler := clientApp.NewGetClientByEmailHandler(repo)
	_, err := handler.Handle("missing@example.com")
	if err == nil {
		t.Fatal("expected error for missing client")
	}
}
