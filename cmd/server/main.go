package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"

	clientApp "github.com/mundoinvest/client/application"
	clientHTTP "github.com/mundoinvest/client/infrastructure/http"
	clientPersistence "github.com/mundoinvest/client/infrastructure/persistence"

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

	clientRepo := clientPersistence.NewSQLiteRepository(db)
	webhookEventRepo := webhookPersistence.NewSQLiteEventRepository(db)

	createClientHandler := clientApp.NewCreateClientHandler(clientRepo, pipefyClient)
	getClientHandler := clientApp.NewGetClientByEmailHandler(clientRepo)

	processCardHandler := webhookApp.NewProcessCardUpdatedHandler(
		webhookEventRepo,
		getClientHandler,
		clientRepo,
		pipefyClient,
	)

	clientHTTPHandler := clientHTTP.NewHandler(createClientHandler)
	webhookHTTPHandler := webhookHTTP.NewHandler(processCardHandler)

	r := gin.Default()
	r.POST("/clientes", clientHTTPHandler.Create)
	r.POST("/webhooks/pipefy/card-updated", webhookHTTPHandler.CardUpdated)

	log.Println("server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	clientRepo := clientPersistence.NewSQLiteRepository(db)
	if err := clientRepo.Migrate(); err != nil {
		return err
	}

	webhookRepo := webhookPersistence.NewSQLiteEventRepository(db)
	if err := webhookRepo.Migrate(); err != nil {
		return err
	}

	return nil
}
