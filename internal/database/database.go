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
	CREATE TABLE IF NOT EXISTS organizations (
		id INTEGER PRIMARY KEY,
		title TEXT,
		last_scraped DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		datetime_start DATETIME NOT NULL,
		datetime_end DATETIME,
		url TEXT NOT NULL UNIQUE,
		location TEXT,
		scraper TEXT DEFAULT 'website',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (organization_id) REFERENCES organizations(id)
	);

	CREATE INDEX IF NOT EXISTS idx_events_org_date ON events(organization_id, datetime_start);
	CREATE INDEX IF NOT EXISTS idx_events_date ON events(datetime_start);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	migrationSQL := `
	ALTER TABLE organizations ADD COLUMN title TEXT;
	`
	db.Exec(migrationSQL)

	return nil
}
