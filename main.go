package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/romanzipp/linke-calendar/internal/config"
	"github.com/romanzipp/linke-calendar/internal/database"
	"github.com/romanzipp/linke-calendar/internal/handlers"
	"github.com/romanzipp/linke-calendar/internal/scraper"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "calendar.db"
	}

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	scheduler := scraper.NewScheduler(db, cfg)
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scraper scheduler: %v", err)
	}

	h, err := handlers.New(db, cfg)
	if err != nil {
		log.Fatalf("Failed to create handlers: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Get("/health", h.Health)
	r.Get("/cal/{siteID}", h.Calendar)
	r.Get("/event/{eventID}", h.EventDetail)

	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	addr := cfg.GetServerAddress()
	log.Printf("Starting server on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
