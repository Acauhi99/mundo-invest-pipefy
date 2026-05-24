package webhook

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS eventos_processados (
		event_id TEXT PRIMARY KEY,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) IsEventProcessed(eventID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM eventos_processados WHERE event_id = ?`, eventID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) MarkEventProcessed(eventID string) error {
	_, err := r.db.Exec(`INSERT INTO eventos_processados (event_id) VALUES (?)`, eventID)
	return err
}
