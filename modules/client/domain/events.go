package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ClientCreated struct {
	ClientID  int64     `json:"client_id"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ClientCreated) EventName() string     { return "client.created" }
func (e ClientCreated) OccurredAt() time.Time { return e.Timestamp }

type ClientProcessed struct {
	Email     string    `json:"email"`
	Priority  string    `json:"priority"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ClientProcessed) EventName() string     { return "client.processed" }
func (e ClientProcessed) OccurredAt() time.Time { return e.Timestamp }
