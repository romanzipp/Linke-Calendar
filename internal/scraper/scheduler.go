package scraper

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/romanzipp/linke-calendar/internal/config"
	"github.com/romanzipp/linke-calendar/internal/database"
)

type Scheduler struct {
	scraper *Scraper
	cron    *cron.Cron
	config  *config.Config
}

func NewScheduler(db *database.DB, cfg *config.Config) *Scheduler {
	return &Scheduler{
		scraper: New(db, cfg),
		cron:    cron.New(),
		config:  cfg,
	}
}

func (s *Scheduler) Start() error {
	interval := s.config.GetScraperInterval()

	log.Printf("Starting scraper scheduler with interval: %v", interval)

	go func() {
		log.Println("Running initial scrape")
		if err := s.scraper.ScrapeAll(); err != nil {
			log.Printf("Initial scrape failed: %v", err)
		}
	}()

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			log.Println("Running scheduled scrape")
			if err := s.scraper.ScrapeAll(); err != nil {
				log.Printf("Scheduled scrape failed: %v", err)
			}
		}
	}()

	return nil
}

func (s *Scheduler) ScrapeNow() error {
	return s.scraper.ScrapeAll()
}
