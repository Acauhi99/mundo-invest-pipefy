package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"

	"github.com/mundoinvest/client-management/internal/cliente"
	"github.com/mundoinvest/client-management/internal/pipefy"
	"github.com/mundoinvest/client-management/internal/webhook"
)

func main() {
	db, err := sql.Open("sqlite", "file:mundoinvest.db?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("erro ao abrir banco: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatalf("erro ao executar migrations: %v", err)
	}

	pipefyClient := pipefy.NewClient()

	clienteRepo := cliente.NewRepository(db)
	webhookEventRepo := webhook.NewRepository(db)

	clienteSvc := cliente.NewService(clienteRepo, pipefyClient)
	webhookSvc := webhook.NewService(webhookEventRepo, clienteRepo, pipefyClient)

	clienteHandler := cliente.NewHandler(clienteSvc)
	webhookHandler := webhook.NewHandler(webhookSvc)

	r := gin.Default()
	r.POST("/clientes", clienteHandler.Criar)
	r.POST("/webhooks/pipefy/card-updated", webhookHandler.CardUpdated)

	log.Println("servidor iniciado em :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	clienteRepo := cliente.NewRepository(db)
	if err := clienteRepo.Migrate(); err != nil {
		return err
	}

	webhookRepo := webhook.NewRepository(db)
	if err := webhookRepo.Migrate(); err != nil {
		return err
	}

	return nil
}
