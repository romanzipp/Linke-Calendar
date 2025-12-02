package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/go-chi/chi/v5"
	"github.com/romanzipp/linke-calendar/internal/calendar"
	"github.com/romanzipp/linke-calendar/internal/database"
)

type Scraper interface {
	ScrapeOrganization(orgID int) error
}

type Handler struct {
	db        *database.DB
	scraper   Scraper
	templates *template.Template
}

func New(db *database.DB, scraper Scraper) (*Handler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:        db,
		scraper:   scraper,
		templates: tmpl,
	}, nil
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Calendar(w http.ResponseWriter, r *http.Request) {
	orgStr := chi.URLParam(r, "org")
	orgID, err := strconv.Atoi(orgStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	hasEvents, err := h.db.HasEventsForOrganization(orgID)
	if err != nil {
		log.Printf("Failed to check events for organization %d: %v", orgID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !hasEvents {
		log.Printf("No events for organization %d, scraping synchronously", orgID)
		if err := h.scraper.ScrapeOrganization(orgID); err != nil {
			log.Printf("Failed to scrape organization %d: %v", orgID, err)
			http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
			return
		}
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	now := time.Now()
	year := now.Year()
	month := now.Month()

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = time.Month(m)
		}
	}

	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	events, err := h.db.GetEventsByOrganizationInRange(orgID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	org, err := h.db.GetOrganization(orgID)
	if err != nil {
		log.Printf("Failed to get organization %d: %v", orgID, err)
	}

	cal := calendar.Generate(year, month, events)

	data := struct {
		OrganizationID    int
		OrganizationTitle string
		Calendar          *calendar.Month
		Year              int
		Month             int
		PrevYear          int
		PrevMonth         int
		NextYear          int
		NextMonth         int
	}{
		OrganizationID:    orgID,
		OrganizationTitle: getOrganizationTitle(org),
		Calendar:          cal,
		Year:              year,
		Month:             int(month),
	}

	prevMonth := month - 1
	prevYear := year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear--
	}
	data.PrevYear = prevYear
	data.PrevMonth = int(prevMonth)

	nextMonth := month + 1
	nextYear := year
	if nextMonth == 13 {
		nextMonth = 1
		nextYear++
	}
	data.NextYear = nextYear
	data.NextMonth = int(nextMonth)

	if r.Header.Get("HX-Request") == "true" {
		if err := h.templates.ExecuteTemplate(w, "calendar-content", data); err != nil {
			log.Printf("Failed to render calendar content: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := h.templates.ExecuteTemplate(w, "calendar.html", data); err != nil {
		log.Printf("Failed to render calendar: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) EventDetail(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := h.db.GetEvent(eventID)
	if err != nil {
		log.Printf("Failed to get event %d: %v", eventID, err)
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	data := struct {
		Event *database.Event
	}{
		Event: event,
	}

	if err := h.templates.ExecuteTemplate(w, "event-modal", data); err != nil {
		log.Printf("Failed to render event modal: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orgStr := chi.URLParam(r, "org")
	orgID, err := strconv.Atoi(orgStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	hasEvents, err := h.db.HasEventsForOrganization(orgID)
	if err != nil {
		log.Printf("Failed to check events for organization %d: %v", orgID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !hasEvents {
		log.Printf("No events for organization %d, scraping synchronously", orgID)
		if err := h.scraper.ScrapeOrganization(orgID); err != nil {
			log.Printf("Failed to scrape organization %d: %v", orgID, err)
			http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
			return
		}
	}

	org, err := h.db.GetOrganization(orgID)
	if err != nil {
		log.Printf("Failed to get organization %d: %v", orgID, err)
	}

	color := r.URL.Query().Get("color")

	events, err := h.db.GetAllUpcomingEventsByOrganization(orgID)
	if err != nil {
		log.Printf("Failed to get upcoming events: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		OrganizationID    int
		OrganizationTitle string
		Events            []*database.Event
		Color             string
	}{
		OrganizationID:    orgID,
		OrganizationTitle: getOrganizationTitle(org),
		Events:            events,
		Color:             color,
	}

	if err := h.templates.ExecuteTemplate(w, "list.html", data); err != nil {
		log.Printf("Failed to render list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) ICalendar(w http.ResponseWriter, r *http.Request) {
	orgStr := chi.URLParam(r, "org")
	orgID, err := strconv.Atoi(orgStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	hasEvents, err := h.db.HasEventsForOrganization(orgID)
	if err != nil {
		log.Printf("Failed to check events for organization %d: %v", orgID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !hasEvents {
		log.Printf("No events for organization %d, scraping synchronously", orgID)
		if err := h.scraper.ScrapeOrganization(orgID); err != nil {
			log.Printf("Failed to scrape organization %d: %v", orgID, err)
			http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
			return
		}
	}

	org, err := h.db.GetOrganization(orgID)
	if err != nil {
		log.Printf("Failed to get organization %d: %v", orgID, err)
	}

	title := getOrganizationTitle(org)

	events, err := h.db.GetEventsByOrganization(orgID)
	if err != nil {
		log.Printf("Failed to get events for organization %d: %v", orgID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Printf("Failed to load Berlin timezone: %v", err)
		berlin = time.UTC
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetName(title)
	cal.SetDescription(fmt.Sprintf("Events calendar for %s", title))
	cal.SetXPublishedTTL("PT1H")
	cal.SetTimezoneId("Europe/Berlin")

	for _, event := range events {
		icalEvent := cal.AddEvent(fmt.Sprintf("%d-%d@linke-calendar", orgID, event.ID))
		icalEvent.SetCreatedTime(event.CreatedAt)
		icalEvent.SetModifiedAt(event.UpdatedAt)
		icalEvent.SetStartAt(reinterpretTimeInLocation(event.DatetimeStart, berlin))

		if event.DatetimeEnd.Valid {
			duration := event.DatetimeEnd.Time.Sub(event.DatetimeStart)
			if duration > 4*24*time.Hour {
				icalEvent.SetEndAt(reinterpretTimeInLocation(event.DatetimeStart, berlin).Add(1 * time.Hour))
			} else {
				icalEvent.SetEndAt(reinterpretTimeInLocation(event.DatetimeEnd.Time, berlin))
			}
		} else {
			icalEvent.SetEndAt(reinterpretTimeInLocation(event.DatetimeStart, berlin).Add(1 * time.Hour))
		}

		icalEvent.SetSummary(event.Title)

		if event.Description.Valid {
			icalEvent.SetDescription(event.Description.String)
		}

		if event.Location.Valid {
			icalEvent.SetLocation(event.Location.String)
		}

		if event.URL != "" {
			icalEvent.SetURL(event.URL)
		}
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%d.ics\"", orgID))

	if err := cal.SerializeTo(w); err != nil {
		log.Printf("Failed to serialize iCal: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func reinterpretTimeInLocation(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func getOrganizationTitle(org *database.Organization) string {
	if org != nil && org.Title.Valid {
		return org.Title.String
	}
	if org != nil {
		return fmt.Sprintf("Organization %d", org.ID)
	}
	return "Unknown Organization"
}
