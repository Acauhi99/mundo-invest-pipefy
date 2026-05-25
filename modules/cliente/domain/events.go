package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ClienteCriado struct {
	ClienteID int64     `json:"cliente_id"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ClienteCriado) EventName() string     { return "cliente.criado" }
func (e ClienteCriado) OccurredAt() time.Time { return e.Timestamp }

type ClienteProcessado struct {
	Email      string    `json:"email"`
	Prioridade string    `json:"prioridade"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e ClienteProcessado) EventName() string     { return "cliente.processado" }
func (e ClienteProcessado) OccurredAt() time.Time { return e.Timestamp }
