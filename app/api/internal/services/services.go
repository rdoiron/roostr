// Package services contains business logic for Roostr.
// Services coordinate between handlers, database, and relay control.
package services

import (
	"github.com/roostr/roostr/app/api/internal/db"
)

// Services holds all application services.
type Services struct {
	Deletion  *DeletionService
	Retention *RetentionService
}

// New creates a new Services instance with all services initialized.
func New(database *db.DB) *Services {
	deletion := NewDeletionService(database)
	retention := NewRetentionService(database, deletion)

	return &Services{
		Deletion:  deletion,
		Retention: retention,
	}
}

// Start starts all background services.
func (s *Services) Start() {
	s.Retention.Start()
}

// Stop stops all background services gracefully.
func (s *Services) Stop() {
	s.Retention.Stop()
}
