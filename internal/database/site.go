package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Site struct {
	ID          string
	Name        string
	URL         string
	LastScraped sql.NullTime
	CreatedAt   time.Time
}

func (db *DB) CreateSite(site *Site) error {
	query := `INSERT INTO sites (id, name, url) VALUES (?, ?, ?)`
	_, err := db.Exec(query, site.ID, site.Name, site.URL)
	if err != nil {
		return fmt.Errorf("failed to create site: %w", err)
	}
	return nil
}

func (db *DB) GetSite(id string) (*Site, error) {
	query := `SELECT id, name, url, last_scraped, created_at FROM sites WHERE id = ?`
	var site Site
	err := db.QueryRow(query, id).Scan(
		&site.ID,
		&site.Name,
		&site.URL,
		&site.LastScraped,
		&site.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get site: %w", err)
	}
	return &site, nil
}

func (db *DB) GetAllSites() ([]*Site, error) {
	query := `SELECT id, name, url, last_scraped, created_at FROM sites ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sites: %w", err)
	}
	defer rows.Close()

	var sites []*Site
	for rows.Next() {
		var site Site
		if err := rows.Scan(
			&site.ID,
			&site.Name,
			&site.URL,
			&site.LastScraped,
			&site.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}
		sites = append(sites, &site)
	}

	return sites, nil
}

func (db *DB) UpdateSiteLastScraped(id string, t time.Time) error {
	query := `UPDATE sites SET last_scraped = ? WHERE id = ?`
	_, err := db.Exec(query, t, id)
	if err != nil {
		return fmt.Errorf("failed to update site last_scraped: %w", err)
	}
	return nil
}

func (db *DB) UpsertSite(site *Site) error {
	query := `
		INSERT INTO sites (id, name, url)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			url = excluded.url
	`
	_, err := db.Exec(query, site.ID, site.Name, site.URL)
	if err != nil {
		return fmt.Errorf("failed to upsert site: %w", err)
	}
	return nil
}
