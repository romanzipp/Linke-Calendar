package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Organization struct {
	ID          int
	LastScraped sql.NullTime
	CreatedAt   time.Time
}

func (db *DB) CreateOrganization(org *Organization) error {
	query := `INSERT INTO organizations (id) VALUES (?)`
	_, err := db.Exec(query, org.ID)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	return nil
}

func (db *DB) GetOrganization(id int) (*Organization, error) {
	query := `SELECT id, last_scraped, created_at FROM organizations WHERE id = ?`
	var org Organization
	err := db.QueryRow(query, id).Scan(
		&org.ID,
		&org.LastScraped,
		&org.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &org, nil
}

func (db *DB) GetAllOrganizations() ([]*Organization, error) {
	query := `SELECT id, last_scraped, created_at FROM organizations ORDER BY id`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(
			&org.ID,
			&org.LastScraped,
			&org.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, &org)
	}

	return orgs, nil
}

func (db *DB) GetDistinctOrganizationsFromEvents() ([]int, error) {
	query := `SELECT DISTINCT organization_id FROM events ORDER BY organization_id`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct organizations: %w", err)
	}
	defer rows.Close()

	var orgIDs []int
	for rows.Next() {
		var orgID int
		if err := rows.Scan(&orgID); err != nil {
			return nil, fmt.Errorf("failed to scan organization id: %w", err)
		}
		orgIDs = append(orgIDs, orgID)
	}

	return orgIDs, nil
}

func (db *DB) UpdateOrganizationLastScraped(id int, t time.Time) error {
	query := `UPDATE organizations SET last_scraped = ? WHERE id = ?`
	_, err := db.Exec(query, t, id)
	if err != nil {
		return fmt.Errorf("failed to update organization last_scraped: %w", err)
	}
	return nil
}

func (db *DB) UpsertOrganization(org *Organization) error {
	query := `
		INSERT INTO organizations (id)
		VALUES (?)
		ON CONFLICT(id) DO UPDATE SET
			last_scraped = COALESCE(organizations.last_scraped, excluded.last_scraped)
	`
	_, err := db.Exec(query, org.ID)
	if err != nil {
		return fmt.Errorf("failed to upsert organization: %w", err)
	}
	return nil
}

func (db *DB) HasEventsForOrganization(orgID int) (bool, error) {
	query := `SELECT COUNT(*) FROM events WHERE organization_id = ?`
	var count int
	err := db.QueryRow(query, orgID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check events for organization: %w", err)
	}
	return count > 0, nil
}
