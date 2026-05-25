package http_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"

	clienteApp "github.com/mundoinvest/cliente/application"
	"github.com/mundoinvest/cliente/domain"
	clientePersistence "github.com/mundoinvest/cliente/infrastructure/persistence"
	"github.com/mundoinvest/pipefy"
	"github.com/mundoinvest/shared"
	webhookApp "github.com/mundoinvest/webhook/application"
	webhookHTTP "github.com/mundoinvest/webhook/infrastructure/http"
	webhookPersistence "github.com/mundoinvest/webhook/infrastructure/persistence"
)

func setupWebhookTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	if err := clienteRepo.Migrate(); err != nil {
		t.Fatalf("failed to migrate clientes: %v", err)
	}
	webhookRepo := webhookPersistence.NewSQLiteEventRepository(db)
	if err := webhookRepo.Migrate(); err != nil {
		t.Fatalf("failed to migrate webhooks: %v", err)
	}
	return db
}

func setupWebhookRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	webhookEventRepo := webhookPersistence.NewSQLiteEventRepository(db)
	pipefyClient := pipefy.NewClient()

	clienteQry := clienteApp.NewObterClientePorEmailHandler(clienteRepo)
	handler := webhookApp.NewProcessarCardUpdatedHandler(webhookEventRepo, clienteQry, clienteRepo, pipefyClient)
	httpHandler := webhookHTTP.NewHandler(handler)
	r.POST("/webhooks/pipefy/card-updated", httpHandler.CardUpdated)
	return r
}

func criarClienteParaTeste(db *sql.DB, nome, email string, patrimonio float64) {
	repo := clientePersistence.NewSQLiteRepository(db)
	c := &domain.Cliente{
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
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !apiResp.Success {
		t.Fatalf("expected success, got error: %+v", apiResp)
	}

	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	c, err := clienteRepo.FindByEmail("joao.silva@example.com")
	if err != nil {
		t.Fatalf("failed to find client: %v", err)
	}
	if c.Status != "Processado" {
		t.Errorf("expected status 'Processado', got '%s'", c.Status)
	}
	if c.Prioridade != "prioridade_alta" {
		t.Errorf("expected priority 'prioridade_alta', got '%s'", c.Prioridade)
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
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !apiResp.Success {
		t.Fatalf("expected success, got error: %+v", apiResp)
	}

	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	c, err := clienteRepo.FindByEmail("maria.souza@example.com")
	if err != nil {
		t.Fatalf("failed to find client: %v", err)
	}
	if c.Prioridade != "prioridade_normal" {
		t.Errorf("expected priority 'prioridade_normal', got '%s'", c.Prioridade)
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
		t.Fatalf("first request expected %d, got %d: %s", http.StatusOK, w1.Code, w1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected status %d for duplicate event, got %d: %s", http.StatusConflict, w2.Code, w2.Body.String())
	}

	var apiResp struct {
		Success bool             `json:"success"`
		Error   *shared.APIError `json:"error"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if apiResp.Success {
		t.Error("expected failure for duplicate event")
	}
	if apiResp.Error.Code != "EVENT_ALREADY_PROCESSED" {
		t.Errorf("expected code EVENT_ALREADY_PROCESSED, got '%s'", apiResp.Error.Code)
	}
}

func TestWebhook_ClienteNaoEncontrado(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer db.Close()

	router := setupWebhookRouter(db)

	payload := map[string]interface{}{
		"event_id":      "evt_004",
		"card_id":       "card_999",
		"cliente_email": "inexistente@example.com",
		"timestamp":     "2026-05-18T12:00:00Z",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for missing client, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool             `json:"success"`
		Error   *shared.APIError `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if apiResp.Success {
		t.Error("expected failure for missing client")
	}
	if apiResp.Error.Code != "CLIENT_NOT_FOUND" {
		t.Errorf("expected code CLIENT_NOT_FOUND, got '%s'", apiResp.Error.Code)
	}
}

func TestWebhook_PatrimonioExatamente200k_PrioridadeAlta(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer db.Close()
	criarClienteParaTeste(db, "Carlos Silva", "carlos.silva@example.com", 200000)

	router := setupWebhookRouter(db)

	payload := map[string]interface{}{
		"event_id":      "evt_005",
		"card_id":       "card_200k",
		"cliente_email": "carlos.silva@example.com",
		"timestamp":     "2026-05-18T12:00:00Z",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/pipefy/card-updated", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !apiResp.Success {
		t.Fatalf("expected success, got error: %+v", apiResp)
	}

	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	c, err := clienteRepo.FindByEmail("carlos.silva@example.com")
	if err != nil {
		t.Fatalf("failed to find client: %v", err)
	}
	if c.Prioridade != "prioridade_alta" {
		t.Errorf("expected priority 'prioridade_alta' for patrimonio=200000, got '%s'", c.Prioridade)
	}
}
