package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/romanzipp/linke-calendar/internal/calendar"
	"github.com/romanzipp/linke-calendar/internal/config"
	"github.com/romanzipp/linke-calendar/internal/database"
)

type Handler struct {
	db        *database.DB
	config    *config.Config
	templates *template.Template
}

func New(db *database.DB, cfg *config.Config) (*Handler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:        db,
		config:    cfg,
		templates: tmpl,
	}, nil
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Calendar(w http.ResponseWriter, r *http.Request) {
	siteID := chi.URLParam(r, "siteID")

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

	site, err := h.db.GetSite(siteID)
	if err != nil {
		log.Printf("Failed to get site %s: %v", siteID, err)
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	events, err := h.db.GetEventsBySiteInRange(siteID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cal := calendar.Generate(year, month, events)

	data := struct {
		Site     *database.Site
		Calendar *calendar.Month
		Year     int
		Month    int
		PrevYear int
		PrevMonth int
		NextYear int
		NextMonth int
	}{
		Site:     site,
		Calendar: cal,
		Year:     year,
		Month:    int(month),
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
