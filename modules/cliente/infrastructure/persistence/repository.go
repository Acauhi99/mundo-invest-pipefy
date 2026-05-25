package persistence

import (
	"database/sql"
	"time"

	"github.com/mundoinvest/cliente/domain"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Migrate() error {
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

func (r *SQLiteRepository) Create(c *domain.Cliente) error {
	query := `INSERT INTO clientes (nome, email, tipo_solicitacao, valor_patrimonio, status, prioridade, card_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`
	return r.db.QueryRow(query, c.Nome, c.Email, c.TipoSolicitacao, c.ValorPatrimonio, c.Status, c.Prioridade, c.CardID).Scan(&c.ID, &c.CreatedAt)
}

func (r *SQLiteRepository) FindByEmail(email string) (*domain.Cliente, error) {
	query := `SELECT id, nome, email, tipo_solicitacao, valor_patrimonio, status, prioridade, card_id, created_at
		FROM clientes WHERE email = ?`
	c := &domain.Cliente{}
	err := r.db.QueryRow(query, email).Scan(
		&c.ID, &c.Nome, &c.Email, &c.TipoSolicitacao, &c.ValorPatrimonio,
		&c.Status, &c.Prioridade, &c.CardID, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *SQLiteRepository) UpdateStatusAndPriority(email, status, prioridade string) error {
	query := `UPDATE clientes SET status = ?, prioridade = ?, updated_at = ? WHERE email = ?`
	_, err := r.db.Exec(query, status, prioridade, time.Now(), email)
	return err
}

func (r *SQLiteRepository) UpdateCardID(email, cardID string) error {
	_, err := r.db.Exec(`UPDATE clientes SET card_id = ?, updated_at = ? WHERE email = ?`, cardID, time.Now(), email)
	return err
}
