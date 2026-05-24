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
)

func setupClienteTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("erro ao abrir banco em memória: %v", err)
	}
	repo := cliente.NewRepository(db)
	if err := repo.Migrate(); err != nil {
		t.Fatalf("erro ao migrar: %v", err)
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
		t.Errorf("esperado status %d, recebido %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp cliente.Cliente
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	if resp.Nome != "João Silva" {
		t.Errorf("esperado nome 'João Silva', recebido '%s'", resp.Nome)
	}
	if resp.Email != "joao.silva@example.com" {
		t.Errorf("esperado email 'joao.silva@example.com', recebido '%s'", resp.Email)
	}
	if resp.Status != "Aguardando Análise" {
		t.Errorf("esperado status 'Aguardando Análise', recebido '%s'", resp.Status)
	}
	if resp.ID == 0 {
		t.Error("esperado ID > 0")
	}

	repo := cliente.NewRepository(db)
	saved, err := repo.FindByEmail("joao.silva@example.com")
	if err != nil {
		t.Fatalf("erro ao buscar cliente no banco: %v", err)
	}
	if saved.Nome != "João Silva" {
		t.Errorf("cliente não foi persistido corretamente")
	}
}
