package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/romanzipp/linke-calendar/internal/config"
	"github.com/romanzipp/linke-calendar/internal/database"
)

type Scraper struct {
	db     *database.DB
	config *config.Config
	client *http.Client
}

func New(db *database.DB, cfg *config.Config) *Scraper {
	return &Scraper{
		db:     db,
		config: cfg,
		client: &http.Client{
			Timeout: cfg.GetScraperTimeout(),
		},
	}
}

func (s *Scraper) ScrapeAll() error {
	log.Println("Starting scrape of all sites")

	for _, site := range s.config.Sites {
		if err := s.ScrapeSite(site); err != nil {
			log.Printf("Error scraping site %s: %v", site.ID, err)
			continue
		}
	}

	log.Println("Completed scrape of all sites")
	return nil
}

func (s *Scraper) ScrapeSite(site config.Site) error {
	log.Printf("Scraping site: %s (%s)", site.Name, site.ID)

	if err := s.db.UpsertSite(&database.Site{
		ID:   site.ID,
		Name: site.Name,
		URL:  site.URL,
	}); err != nil {
		return fmt.Errorf("failed to upsert site: %w", err)
	}

	maxPages := s.config.GetScraperMaxPages() + 1
	totalEvents := 0

	for page := 1; page < maxPages; page++ {
		pageURL := s.buildPageURL(site.URL, page)
		log.Printf("Fetching page %d: %s", page, pageURL)

		html, err := s.fetchPage(pageURL)
		if err != nil {
			log.Printf("Failed to fetch page %d: %v", page, err)
			break
		}

		events, err := ParseHTML(html, pageURL)
		if err != nil {
			log.Printf("Failed to parse HTML on page %d: %v", page, err)
			break
		}

		if len(events) == 0 {
			log.Printf("No events found on page %d, stopping", page)
			break
		}

		for _, event := range events {
			dbEvent := &database.Event{
				SiteID:        site.ID,
				Title:         event.Title,
				Description:   toNullString(event.Description),
				DatetimeStart: event.DateTime,
				DatetimeEnd:   sql.NullTime{},
				URL:           event.URL,
				Location:      toNullString(event.Location),
				Typo3URL:      toNullString(event.URL),
			}

			if err := s.db.UpsertEvent(dbEvent); err != nil {
				log.Printf("Failed to upsert event %s: %v", event.Title, err)
				continue
			}
			totalEvents++
		}

		time.Sleep(1 * time.Second)
	}

	if err := s.db.UpdateSiteLastScraped(site.ID, time.Now()); err != nil {
		log.Printf("Failed to update last_scraped for site %s: %v", site.ID, err)
	}

	log.Printf("Scraped %d events from site %s", totalEvents, site.ID)
	return nil
}

func (s *Scraper) fetchPage(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:145.0) Gecko/20100101 Firefox/145.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("HTTP error response preview: %s", preview)
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	return string(body), nil
}

func (s *Scraper) buildPageURL(urlTemplate string, page int) string {
	return strings.ReplaceAll(urlTemplate, "{page}", fmt.Sprintf("%d", page))
}
