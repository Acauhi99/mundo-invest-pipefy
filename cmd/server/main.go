package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"

	clienteApp "github.com/mundoinvest/cliente/application"
	clienteHTTP "github.com/mundoinvest/cliente/infrastructure/http"
	clientePersistence "github.com/mundoinvest/cliente/infrastructure/persistence"

	webhookApp "github.com/mundoinvest/webhook/application"
	webhookHTTP "github.com/mundoinvest/webhook/infrastructure/http"
	webhookPersistence "github.com/mundoinvest/webhook/infrastructure/persistence"

	"github.com/mundoinvest/pipefy"
)

func main() {
	db, err := sql.Open("sqlite", "file:mundoinvest.db?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	pipefyClient := pipefy.NewClient()

	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	webhookEventRepo := webhookPersistence.NewSQLiteEventRepository(db)

	criarClienteHandler := clienteApp.NewCriarClienteHandler(clienteRepo, pipefyClient)
	obterClienteHandler := clienteApp.NewObterClientePorEmailHandler(clienteRepo)

	processarCardHandler := webhookApp.NewProcessarCardUpdatedHandler(
		webhookEventRepo,
		obterClienteHandler,
		clienteRepo,
		pipefyClient,
	)

	clienteHTTPHandler := clienteHTTP.NewHandler(criarClienteHandler)
	webhookHTTPHandler := webhookHTTP.NewHandler(processarCardHandler)

	r := gin.Default()
	r.POST("/clientes", clienteHTTPHandler.Criar)
	r.POST("/webhooks/pipefy/card-updated", webhookHTTPHandler.CardUpdated)

	log.Println("server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	clienteRepo := clientePersistence.NewSQLiteRepository(db)
	if err := clienteRepo.Migrate(); err != nil {
		return err
	}

	webhookRepo := webhookPersistence.NewSQLiteEventRepository(db)
	if err := webhookRepo.Migrate(); err != nil {
		return err
	}

	return nil
}
