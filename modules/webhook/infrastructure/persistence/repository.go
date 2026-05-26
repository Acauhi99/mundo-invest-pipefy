package persistence

import "database/sql"

type SQLiteEventRepository struct {
	db *sql.DB
}

func NewSQLiteEventRepository(db *sql.DB) *SQLiteEventRepository {
	return &SQLiteEventRepository{db: db}
}

func (r *SQLiteEventRepository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS processed_events (
		event_id TEXT PRIMARY KEY,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteEventRepository) IsEventProcessed(eventID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM processed_events WHERE event_id = ?`, eventID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLiteEventRepository) MarkEventProcessed(eventID string) error {
	_, err := r.db.Exec(`INSERT INTO processed_events (event_id) VALUES (?)`, eventID)
	return err
}
