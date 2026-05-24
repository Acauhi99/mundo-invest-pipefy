package webhook

import "time"

type CardUpdatedInput struct {
	EventID      string `json:"event_id" binding:"required"`
	CardID       string `json:"card_id" binding:"required"`
	ClienteEmail string `json:"cliente_email" binding:"required,email"`
	Timestamp    string `json:"timestamp" binding:"required"`
}

type ProcessedEvent struct {
	EventID     string    `json:"event_id"`
	ProcessedAt time.Time `json:"processed_at"`
}
