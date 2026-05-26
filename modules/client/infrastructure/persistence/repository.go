package persistence

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mundoinvest/client/domain"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		request_type TEXT NOT NULL,
		net_worth REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'Aguardando Análise',
		priority TEXT DEFAULT '',
		card_id TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME
	)`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(c *domain.Client) error {
	query := `INSERT INTO clients (name, email, request_type, net_worth, status, priority, card_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`
	return r.db.QueryRow(query, c.Name, c.Email, c.RequestType, c.NetWorth, c.Status, c.Priority, c.CardID).Scan(&c.ID, &c.CreatedAt)
}

func (r *SQLiteRepository) FindByEmail(email string) (*domain.Client, error) {
	query := `SELECT id, name, email, request_type, net_worth, status, priority, card_id, created_at
		FROM clients WHERE email = ?`
	c := &domain.Client{}
	err := r.db.QueryRow(query, email).Scan(
		&c.ID, &c.Name, &c.Email, &c.RequestType, &c.NetWorth,
		&c.Status, &c.Priority, &c.CardID, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrClientNotFound
		}
		return nil, err
	}
	return c, nil
}

func (r *SQLiteRepository) UpdateStatusAndPriority(email, status, priority string) error {
	query := `UPDATE clients SET status = ?, priority = ?, updated_at = ? WHERE email = ?`
	_, err := r.db.Exec(query, status, priority, time.Now(), email)
	return err
}

func (r *SQLiteRepository) UpdateCardID(email, cardID string) error {
	_, err := r.db.Exec(`UPDATE clients SET card_id = ?, updated_at = ? WHERE email = ?`, cardID, time.Now(), email)
	return err
}
