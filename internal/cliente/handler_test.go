package cliente_test

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
	"github.com/mundoinvest/client-management/internal/response"
)

func setupClienteTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	repo := cliente.NewRepository(db)
	if err := repo.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func setupClienteRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := cliente.NewRepository(db)
	pipefyClient := pipefy.NewClient()
	svc := cliente.NewService(repo, pipefyClient)
	handler := cliente.NewHandler(svc)
	r.POST("/clientes", handler.Criar)
	return r
}

func TestCriarCliente_Sucesso(t *testing.T) {
	db := setupClienteTestDB(t)
	defer db.Close()
	router := setupClienteRouter(db)

	payload := map[string]interface{}{
		"cliente_nome":     "João Silva",
		"cliente_email":    "joao.silva@example.com",
		"tipo_solicitacao": "Atualização cadastral",
		"valor_patrimonio": 250000,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool               `json:"success"`
		Data    cliente.Cliente    `json:"data"`
		Error   *response.APIError `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !apiResp.Success {
		t.Fatalf("expected success, got error: %+v", apiResp.Error)
	}
	resp := apiResp.Data

	if resp.Nome != "João Silva" {
		t.Errorf("expected nome 'João Silva', got '%s'", resp.Nome)
	}
	if resp.Email != "joao.silva@example.com" {
		t.Errorf("expected email 'joao.silva@example.com', got '%s'", resp.Email)
	}
	if resp.Status != "Aguardando Análise" {
		t.Errorf("expected status 'Aguardando Análise', got '%s'", resp.Status)
	}
	if resp.ID == 0 {
		t.Error("expected ID > 0")
	}
	if resp.CreatedAt.IsZero() {
		t.Error("expected created_at set")
	}
	if resp.CardID == "" {
		t.Error("expected card_id set")
	}

	repo := cliente.NewRepository(db)
	saved, err := repo.FindByEmail("joao.silva@example.com")
	if err != nil {
		t.Fatalf("failed to find client in database: %v", err)
	}
	if saved.Nome != "João Silva" {
		t.Errorf("client was not persisted correctly")
	}
}

func TestCriarCliente_EmailInvalido(t *testing.T) {
	db := setupClienteTestDB(t)
	defer db.Close()
	router := setupClienteRouter(db)

	payload := map[string]interface{}{
		"cliente_nome":     "João Silva",
		"cliente_email":    "invalido",
		"tipo_solicitacao": "Atualização cadastral",
		"valor_patrimonio": 250000,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Error   response.APIError `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if apiResp.Success {
		t.Error("expected failure")
	}
	if apiResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got '%s'", apiResp.Error.Code)
	}
}

func TestCriarCliente_CamposObrigatoriosAusentes(t *testing.T) {
	db := setupClienteTestDB(t)
	defer db.Close()
	router := setupClienteRouter(db)

	payload := map[string]interface{}{
		"cliente_nome": "João Silva",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Error   response.APIError `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if apiResp.Success {
		t.Error("expected failure")
	}
	if apiResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got '%s'", apiResp.Error.Code)
	}
}

func TestCriarCliente_ValorPatrimonioInvalido(t *testing.T) {
	db := setupClienteTestDB(t)
	defer db.Close()
	router := setupClienteRouter(db)

	payload := map[string]interface{}{
		"cliente_nome":     "João Silva",
		"cliente_email":    "joao.silva@example.com",
		"tipo_solicitacao": "Atualização cadastral",
		"valor_patrimonio": 0,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Error   response.APIError `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if apiResp.Success {
		t.Error("expected failure")
	}
	if apiResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got '%s'", apiResp.Error.Code)
	}
}
