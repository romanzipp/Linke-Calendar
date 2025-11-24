package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ScrapedEvent struct {
	Title       string
	Description string
	DateTime    time.Time
	URL         string
	Location    string
}

func ParseHTML(htmlContent string, baseURL string) ([]*ScrapedEvent, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var events []*ScrapedEvent

	doc.Find(".termin.post-wrapper").Each(func(i int, s *goquery.Selection) {
		event, err := parseEvent(s, baseURL)
		if err != nil {
			log.Printf("Warning: failed to parse event %d: %v", i, err)
			return
		}
		events = append(events, event)
	})

	return events, nil
}

func parseEvent(s *goquery.Selection, baseURL string) (*ScrapedEvent, error) {
	title := strings.TrimSpace(s.Find("h2.card-title a").Text())
	if title == "" {
		return nil, fmt.Errorf("missing title")
	}

	eventURL, exists := s.Find("h2.card-title a").Attr("href")
	if !exists || eventURL == "" {
		return nil, fmt.Errorf("missing event URL")
	}

	eventURL = resolveURL(baseURL, eventURL)

	datetimeStr, exists := s.Find("time.termin-zeiten-datum").Attr("datetime")
	if !exists || datetimeStr == "" {
		return nil, fmt.Errorf("missing datetime")
	}

	datetime, err := parseDateTime(datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %w", err)
	}

	description := extractDescription(s)
	location := extractLocation(description)

	return &ScrapedEvent{
		Title:       title,
		Description: description,
		DateTime:    datetime,
		URL:         eventURL,
		Location:    location,
	}, nil
}

func parseDateTime(datetimeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, datetimeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", datetimeStr)
}

func extractDescription(s *goquery.Selection) string {
	desc := s.Find("p[itemprop='description']").Text()
	desc = strings.TrimSpace(desc)

	if idx := strings.Index(desc, "Weiterlesen"); idx != -1 {
		desc = strings.TrimSpace(desc[:idx])
	}

	return desc
}

func extractLocation(description string) string {
	lines := strings.Split(description, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.Contains(firstLine, "im ") || strings.Contains(firstLine, "in ") {
			return firstLine
		}
	}
	return ""
}

func resolveURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}

	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	return baseURL.ResolveReference(refURL).String()
}

func (se *ScrapedEvent) ToDBEvent(siteID string) *Event {
	return &Event{
		SiteID:        siteID,
		Title:         se.Title,
		Description:   toNullString(se.Description),
		DatetimeStart: se.DateTime,
		DatetimeEnd:   sql.NullTime{},
		URL:           se.URL,
		Location:      toNullString(se.Location),
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

type Event struct {
	SiteID        string
	Title         string
	Description   sql.NullString
	DatetimeStart time.Time
	DatetimeEnd   sql.NullTime
	URL           string
	Location      sql.NullString
}
