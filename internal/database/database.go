package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) Initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		last_scraped DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		site_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		datetime_start DATETIME NOT NULL,
		datetime_end DATETIME,
		url TEXT NOT NULL UNIQUE,
		location TEXT,
		typo3_url TEXT,
		scraper TEXT DEFAULT 'website',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id)
	);

	CREATE INDEX IF NOT EXISTS idx_events_site_date ON events(site_id, datetime_start);
	CREATE INDEX IF NOT EXISTS idx_events_date ON events(datetime_start);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	migrationSQL := `
	ALTER TABLE events ADD COLUMN scraper TEXT DEFAULT 'website';
	`
	db.Exec(migrationSQL)

	return nil
}
