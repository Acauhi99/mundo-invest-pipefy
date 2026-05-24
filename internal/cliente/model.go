package cliente

import "time"

type Cliente struct {
	ID              int64     `json:"id"`
	Nome            string    `json:"cliente_nome"`
	Email           string    `json:"cliente_email"`
	TipoSolicitacao string    `json:"tipo_solicitacao"`
	ValorPatrimonio float64   `json:"valor_patrimonio"`
	Status          string    `json:"status"`
	Prioridade      string    `json:"prioridade,omitempty"`
	CardID          string    `json:"card_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type CriarClienteInput struct {
	Nome            string  `json:"cliente_nome" binding:"required"`
	Email           string  `json:"cliente_email" binding:"required,email"`
	TipoSolicitacao string  `json:"tipo_solicitacao" binding:"required"`
	ValorPatrimonio float64 `json:"valor_patrimonio" binding:"required,gt=0"`
}
