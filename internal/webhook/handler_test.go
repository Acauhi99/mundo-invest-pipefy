package webhook_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"

	"github.com/mundoinvest/client-management/internal/cliente"
	"github.com/mundoinvest/client-management/internal/pipefy"
	"github.com/mundoinvest/client-management/internal/webhook"
)

func setupWebhookTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("erro ao abrir banco em memória: %v", err)
	}
	clienteRepo := cliente.NewRepository(db)
	if err := clienteRepo.Migrate(); err != nil {
		t.Fatalf("erro ao migrar clientes: %v", err)
	}
	webhookRepo := webhook.NewRepository(db)
	if err := webhookRepo.Migrate(); err != nil {
		t.Fatalf("erro ao migrar webhooks: %v", err)
	}
	return db
}

func setupWebhookRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	clienteRepo := cliente.NewRepository(db)
	webhookEventRepo := webhook.NewRepository(db)
	pipefyClient := pipefy.NewClient()
	svc := webhook.NewService(webhookEventRepo, clienteRepo, pipefyClient)
	handler := webhook.NewHandler(svc)
	r.POST("/webhooks/pipefy/card-updated", handler.CardUpdated)
	return r
}

func criarClienteParaTeste(db *sql.DB, nome, email string, patrimonio float64) {
	repo := cliente.NewRepository(db)
	c := &cliente.Cliente{
		Nome:            nome,
		Email:           email,
		TipoSolicitacao: "Atualização cadastral",
		ValorPatrimonio: patrimonio,
		Status:          "Aguardando Análise",
	}
	if err := repo.Create(c); err != nil {
		panic(err)
	}
}

func TestWebhook_PatrimonioAlto_PrioridadeAlta(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer db.Close()
	criarClienteParaTeste(db, "João Silva", "joao.silva@example.com", 250000)

	router := setupWebhookRouter(db)

	payload := map[string]interface{}{
		"event_id":      "evt_001",
		"card_id":       "card_456",
		"cliente_email": "joao.silva@example.com",
		"timestamp":     "2026-05-18T12:00:00Z",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado status %d, recebido %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	clienteRepo := cliente.NewRepository(db)
	c, err := clienteRepo.FindByEmail("joao.silva@example.com")
	if err != nil {
		t.Fatalf("erro ao buscar cliente: %v", err)
	}
	if c.Status != "Processado" {
		t.Errorf("esperado status 'Processado', recebido '%s'", c.Status)
	}
	if c.Prioridade != "prioridade_alta" {
		t.Errorf("esperado prioridade 'prioridade_alta', recebido '%s'", c.Prioridade)
	}
}

func TestWebhook_PatrimonioBaixo_PrioridadeNormal(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer db.Close()
	criarClienteParaTeste(db, "Maria Souza", "maria.souza@example.com", 50000)

	router := setupWebhookRouter(db)

	payload := map[string]interface{}{
		"event_id":      "evt_002",
		"card_id":       "card_789",
		"cliente_email": "maria.souza@example.com",
		"timestamp":     "2026-05-18T12:00:00Z",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado status %d, recebido %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	clienteRepo := cliente.NewRepository(db)
	c, err := clienteRepo.FindByEmail("maria.souza@example.com")
	if err != nil {
		t.Fatalf("erro ao buscar cliente: %v", err)
	}
	if c.Prioridade != "prioridade_normal" {
		t.Errorf("esperado prioridade 'prioridade_normal', recebido '%s'", c.Prioridade)
	}
}

func TestWebhook_EventoDuplicado_Bloqueado(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer db.Close()
	criarClienteParaTeste(db, "João Silva", "joao.silva@example.com", 300000)

	router := setupWebhookRouter(db)

	payload := map[string]interface{}{
		"event_id":      "evt_003",
		"card_id":       "card_111",
		"cliente_email": "joao.silva@example.com",
		"timestamp":     "2026-05-18T12:00:00Z",
	}
	body, _ := json.Marshal(payload)

	req1 := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("primeira requisição esperada %d, recebida %d: %s", http.StatusOK, w1.Code, w1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("esperado status %d para evento duplicado, recebido %d: %s", http.StatusConflict, w2.Code, w2.Body.String())
	}
}
