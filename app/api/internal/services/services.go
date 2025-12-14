// Package services contains business logic for Roostr.
// Services coordinate between handlers, database, and relay control.
package services

import (
	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/relay"
)

// Services holds all application services.
type Services struct {
	Deletion       *DeletionService
	Retention      *RetentionService
	Sync           *SyncService
	Lightning      *LightningService
	InvoiceMonitor *InvoiceMonitorService
}

// New creates a new Services instance with all services initialized.
// The configMgr and relayCtl parameters are used by InvoiceMonitorService
// to sync the whitelist and reload the relay when payments are confirmed.
func New(database *db.DB, configMgr *relay.ConfigManager, relayCtl *relay.Relay) *Services {
	deletion := NewDeletionService(database)
	retention := NewRetentionService(database, deletion)
	sync := NewSyncService(database)
	lightning := NewLightningService(database)
	invoiceMonitor := NewInvoiceMonitorService(database, lightning, configMgr, relayCtl)

	return &Services{
		Deletion:       deletion,
		Retention:      retention,
		Sync:           sync,
		Lightning:      lightning,
		InvoiceMonitor: invoiceMonitor,
	}
}

// Start starts all background services.
func (s *Services) Start() {
	s.Retention.Start()
	s.InvoiceMonitor.Start()
}

// Stop stops all background services gracefully.
func (s *Services) Stop() {
	s.InvoiceMonitor.Stop()
	s.Retention.Stop()
}
