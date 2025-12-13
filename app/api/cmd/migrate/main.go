package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/roostr/roostr/app/api/internal/db"
)

func main() {
	log.Println("Roostr Database Migration Tool")
	log.Println("===============================")

	// Get paths from environment
	appDBPath := os.Getenv("APP_DB_PATH")
	if appDBPath == "" {
		appDBPath = "data/roostr.db"
	}

	relayDBPath := os.Getenv("RELAY_DB_PATH")
	if relayDBPath == "" {
		relayDBPath = "data/nostr.db"
	}

	// Initialize database (this will create and apply schema if new)
	log.Printf("App database: %s", appDBPath)
	database, err := db.New(relayDBPath, appDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Get current schema version
	version, err := database.GetSchemaVersion()
	if err != nil {
		log.Fatalf("Failed to get schema version: %v", err)
	}
	log.Printf("Current schema version: %d", version)

	// Check for pending migrations
	pending, err := database.GetPendingMigrations()
	if err != nil {
		log.Fatalf("Failed to check pending migrations: %v", err)
	}

	if len(pending) == 0 {
		log.Println("No pending migrations")
	} else {
		log.Printf("Found %d pending migration(s)", len(pending))

		// Run migrations
		ctx := context.Background()
		if err := database.Migrate(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	// Show final status
	newVersion, _ := database.GetSchemaVersion()
	fmt.Println()
	log.Printf("Database is now at schema version %d", newVersion)
	log.Println("Migration complete!")
}
