package domain

import "time"

type Client struct {
	ID          int64     `json:"id"`
	Name        string    `json:"cliente_nome"`
	Email       string    `json:"cliente_email"`
	RequestType string    `json:"tipo_solicitacao"`
	NetWorth    float64   `json:"valor_patrimonio"`
	Status      string    `json:"status"`
	Priority    string    `json:"prioridade,omitempty"`
	CardID      string    `json:"card_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
