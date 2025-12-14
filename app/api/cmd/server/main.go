package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roostr/roostr/app/api/internal/config"
	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/handlers"
	"github.com/roostr/roostr/app/api/internal/relay"
	"github.com/roostr/roostr/app/api/internal/services"
)

func main() {
	log.Println("Starting Roostr API server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	database, err := db.New(cfg.RelayDBPath, cfg.AppDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run any pending migrations
	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	// Initialize config manager for relay config.toml
	var configMgr *relay.ConfigManager
	if cfg.ConfigPath != "" {
		configMgr = relay.NewConfigManager(cfg.ConfigPath)
		log.Printf("Config manager initialized for: %s", cfg.ConfigPath)
	}

	// Initialize relay manager
	var relayMgr *relay.Relay
	if cfg.RelayBinary != "" {
		relayMgr = relay.New(cfg.RelayBinary, cfg.ConfigPath)
		if relayMgr.IsRunning() {
			log.Println("Relay process detected as running")
		} else {
			log.Println("Relay process not detected (will sync config but not reload)")
		}
	}

	// Initialize services
	svc := services.New(database)
	svc.Start()
	defer svc.Stop()
	log.Println("Background services started")

	// Create handler with dependencies
	h := handlers.New(database, cfg, configMgr, relayMgr)

	// Set up router
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Apply middleware
	handler := handlers.Chain(mux,
		handlers.Recover,
		handlers.CORS,
		handlers.Logging,
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Roostr API listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop background services first
	log.Println("Stopping background services...")
	svc.Stop()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
