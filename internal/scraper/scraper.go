package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/romanzipp/linke-calendar/internal/config"
	"github.com/romanzipp/linke-calendar/internal/database"
)

type Scraper struct {
	db     *database.DB
	config *config.Config
}

func New(db *database.DB, cfg *config.Config) *Scraper {
	return &Scraper{
		db:     db,
		config: cfg,
	}
}

func (s *Scraper) ScrapeAll() error {
	log.Println("Starting scrape of all organizations")

	orgIDs, err := s.db.GetDistinctOrganizationsFromEvents()
	if err != nil {
		return fmt.Errorf("failed to get organizations: %w", err)
	}

	for _, orgID := range orgIDs {
		if err := s.ScrapeOrganization(orgID); err != nil {
			log.Printf("Error scraping organization %d: %v", orgID, err)
			continue
		}
	}

	log.Println("Completed scrape of all organizations")
	return nil
}

func (s *Scraper) ScrapeOrganization(orgID int) error {
	log.Printf("Scraping organization: %d", orgID)

	if err := s.db.UpsertOrganization(&database.Organization{
		ID: orgID,
	}); err != nil {
		return fmt.Errorf("failed to upsert organization: %w", err)
	}

	totalEvents := s.scrapeZetkin(orgID)

	if err := s.db.UpdateOrganizationLastScraped(orgID, time.Now()); err != nil {
		log.Printf("Failed to update last_scraped for organization %d: %v", orgID, err)
	}

	log.Printf("Scraped %d total events from organization %d", totalEvents, orgID)
	return nil
}

func (s *Scraper) scrapeZetkin(orgID int) int {
	log.Printf("Fetching Zetkin events for organization ID: %d", orgID)

	client := NewZetkinClient(orgID, s.config.GetScraperTimeout())

	events, err := client.FetchAllEvents()
	if err != nil {
		log.Printf("Failed to fetch Zetkin events: %v", err)
		return 0
	}

	log.Printf("Fetched %d events from Zetkin for organization %d", len(events), orgID)

	totalEvents := 0
	for _, event := range events {
		startTime, err := parseZetkinTime(event.StartTime)
		if err != nil {
			log.Printf("Failed to parse start time for event %s: %v", event.Title, err)
			continue
		}

		endTime, err := parseZetkinTime(event.EndTime)
		if err != nil {
			log.Printf("Failed to parse end time for event %s: %v", event.Title, err)
			continue
		}

		location := ""
		if event.Location != nil {
			location = event.Location.Title
		}

		description := event.InfoText
		if description == "" && event.Activity != nil {
			description = event.Activity.Title
		}
		if event.Contact != nil && event.Contact.Name != "" {
			if description != "" {
				description += "\n\n"
			}
			description += "Kontakt: " + event.Contact.Name
		}

		eventURL := fmt.Sprintf("https://app.zetkin.die-linke.de/o/%d/events/%d", event.Organization.ID, event.ID)

		dbEvent := &database.Event{
			OrganizationID: orgID,
			Title:          event.Title,
			Description:    toNullString(description),
			DatetimeStart:  startTime,
			DatetimeEnd:    sql.NullTime{Time: endTime, Valid: true},
			URL:            eventURL,
			Location:       toNullString(location),
			Scraper:        "zetkin",
		}

		if err := s.db.UpsertEvent(dbEvent); err != nil {
			log.Printf("Failed to upsert Zetkin event %s: %v", event.Title, err)
			continue
		}
		totalEvents++
	}

	log.Printf("Scraped %d events from Zetkin for organization %d", totalEvents, orgID)
	return totalEvents
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
