package cliente

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		tipo_solicitacao TEXT NOT NULL,
		valor_patrimonio REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'Aguardando Análise',
		prioridade TEXT DEFAULT '',
		card_id TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME
	)`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) Create(c *Cliente) error {
	query := `INSERT INTO clientes (nome, email, tipo_solicitacao, valor_patrimonio, status, prioridade, card_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`
	return r.db.QueryRow(query, c.Nome, c.Email, c.TipoSolicitacao, c.ValorPatrimonio, c.Status, c.Prioridade, c.CardID).Scan(&c.ID, &c.CreatedAt)
}

func (r *Repository) FindByEmail(email string) (*Cliente, error) {
	query := `SELECT id, nome, email, tipo_solicitacao, valor_patrimonio, status, prioridade, card_id, created_at
		FROM clientes WHERE email = ?`
	c := &Cliente{}
	err := r.db.QueryRow(query, email).Scan(
		&c.ID, &c.Nome, &c.Email, &c.TipoSolicitacao, &c.ValorPatrimonio,
		&c.Status, &c.Prioridade, &c.CardID, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repository) UpdateStatusAndPriority(email, status, prioridade string) error {
	query := `UPDATE clientes SET status = ?, prioridade = ?, updated_at = ? WHERE email = ?`
	_, err := r.db.Exec(query, status, prioridade, time.Now(), email)
	return err
}

func (r *Repository) UpdateCardID(email, cardID string) error {
	_, err := r.db.Exec(`UPDATE clientes SET card_id = ? WHERE email = ?`, cardID, email)
	return err
}
