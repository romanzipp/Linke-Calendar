package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Event struct {
	ID             int
	OrganizationID int
	Title          string
	Description    sql.NullString
	DatetimeStart  time.Time
	DatetimeEnd    sql.NullTime
	URL            string
	Location       sql.NullString
	Scraper        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (db *DB) CreateEvent(event *Event) error {
	query := `
		INSERT INTO events (
			organization_id, title, description, datetime_start, datetime_end,
			url, location, scraper
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(
		query,
		event.OrganizationID,
		event.Title,
		event.Description,
		event.DatetimeStart,
		event.DatetimeEnd,
		event.URL,
		event.Location,
		event.Scraper,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	event.ID = int(id)
	return nil
}

func (db *DB) GetEvent(id int) (*Event, error) {
	query := `
		SELECT id, organization_id, title, description, datetime_start, datetime_end,
		       url, location, scraper, created_at, updated_at
		FROM events WHERE id = ?
	`
	var event Event
	err := db.QueryRow(query, id).Scan(
		&event.ID,
		&event.OrganizationID,
		&event.Title,
		&event.Description,
		&event.DatetimeStart,
		&event.DatetimeEnd,
		&event.URL,
		&event.Location,
		&event.Scraper,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return &event, nil
}

func (db *DB) GetEventsByOrganization(orgID int) ([]*Event, error) {
	query := `
		SELECT id, organization_id, title, description, datetime_start, datetime_end,
		       url, location, scraper, created_at, updated_at
		FROM events
		WHERE organization_id = ?
		ORDER BY datetime_start ASC
	`
	return db.queryEvents(query, orgID)
}

func (db *DB) GetEventsByOrganizationInRange(orgID int, start, end time.Time) ([]*Event, error) {
	query := `
		SELECT id, organization_id, title, description, datetime_start, datetime_end,
		       url, location, scraper, created_at, updated_at
		FROM events
		WHERE organization_id = ? AND datetime_start >= ? AND datetime_start < ?
		ORDER BY datetime_start ASC
	`
	return db.queryEvents(query, orgID, start, end)
}

func (db *DB) GetUpcomingEventsByOrganization(orgID int, limit int) ([]*Event, error) {
	query := `
		SELECT id, organization_id, title, description, datetime_start, datetime_end,
		       url, location, scraper, created_at, updated_at
		FROM events
		WHERE organization_id = ? AND datetime_start >= datetime('now')
		ORDER BY datetime_start ASC
		LIMIT ?
	`
	return db.queryEvents(query, orgID, limit)
}

func (db *DB) GetAllUpcomingEventsByOrganization(orgID int) ([]*Event, error) {
	query := `
		SELECT id, organization_id, title, description, datetime_start, datetime_end,
		       url, location, scraper, created_at, updated_at
		FROM events
		WHERE organization_id = ? AND datetime_start >= datetime('now')
		ORDER BY datetime_start ASC
	`
	return db.queryEvents(query, orgID)
}

func (db *DB) UpsertEvent(event *Event) error {
	query := `
		INSERT INTO events (
			organization_id, title, description, datetime_start, datetime_end,
			url, location, scraper
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			datetime_start = excluded.datetime_start,
			datetime_end = excluded.datetime_end,
			location = excluded.location,
			scraper = excluded.scraper,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(
		query,
		event.OrganizationID,
		event.Title,
		event.Description,
		event.DatetimeStart,
		event.DatetimeEnd,
		event.URL,
		event.Location,
		event.Scraper,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert event: %w", err)
	}
	return nil
}

func (db *DB) DeleteOldEvents(before time.Time) error {
	query := `DELETE FROM events WHERE datetime_start < ?`
	_, err := db.Exec(query, before)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}
	return nil
}

func (db *DB) queryEvents(query string, args ...interface{}) ([]*Event, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.ID,
			&event.OrganizationID,
			&event.Title,
			&event.Description,
			&event.DatetimeStart,
			&event.DatetimeEnd,
			&event.URL,
			&event.Location,
			&event.Scraper,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}
