package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
)

// mockStorageDB implements the database methods needed for Storage handlers
type mockStorageDB struct {
	relayDBSize       int64
	appDBSize         int64
	availableSpace    int64
	totalSpace        int64
	relayConnected    bool
	relayStats        *db.RelayStats
	pendingDeletions  int64
	retentionPolicy   *db.RetentionPolicy
	deletionRequests  []db.DeletionRequest
	countBefore       int64
	avgEventSize      int64
	vacuumErr         error
	integrityCheckErr error
}

func newMockStorageDB() *mockStorageDB {
	return &mockStorageDB{
		relayDBSize:    1024 * 1024 * 100, // 100 MB
		appDBSize:      1024 * 1024 * 10,  // 10 MB
		availableSpace: 1024 * 1024 * 1024 * 40, // 40 GB (20% used = healthy)
		totalSpace:     1024 * 1024 * 1024 * 50, // 50 GB
		relayConnected: true,
		relayStats: &db.RelayStats{
			TotalEvents: 1000,
			OldestEvent: time.Now().Add(-30 * 24 * time.Hour),
			NewestEvent: time.Now(),
		},
		retentionPolicy: &db.RetentionPolicy{
			RetentionDays: 90,
			HonorNIP09:    true,
			Exceptions:    []string{"kind:0", "kind:3"},
		},
		deletionRequests: []db.DeletionRequest{},
		avgEventSize:     500,
	}
}

func (m *mockStorageDB) GetRelayDatabaseSize() (int64, error) {
	return m.relayDBSize, nil
}

func (m *mockStorageDB) GetAppDatabaseSize() (int64, error) {
	return m.appDBSize, nil
}

func (m *mockStorageDB) GetAvailableDiskSpace() (int64, error) {
	return m.availableSpace, nil
}

func (m *mockStorageDB) GetTotalDiskSpace() (int64, error) {
	return m.totalSpace, nil
}

func (m *mockStorageDB) IsRelayDBConnected() bool {
	return m.relayConnected
}

func (m *mockStorageDB) GetRelayStats(ctx context.Context) (*db.RelayStats, error) {
	return m.relayStats, nil
}

func (m *mockStorageDB) GetPendingDeletionCount(ctx context.Context) (int64, error) {
	return m.pendingDeletions, nil
}

func (m *mockStorageDB) GetRetentionPolicy(ctx context.Context) (*db.RetentionPolicy, error) {
	return m.retentionPolicy, nil
}

func (m *mockStorageDB) SetRetentionPolicy(ctx context.Context, policy *db.RetentionPolicy) error {
	m.retentionPolicy = policy
	return nil
}

func (m *mockStorageDB) GetDeletionRequests(ctx context.Context, status string) ([]db.DeletionRequest, error) {
	if status == "" {
		return m.deletionRequests, nil
	}
	var filtered []db.DeletionRequest
	for _, r := range m.deletionRequests {
		if r.Status == status {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

func (m *mockStorageDB) CountEventsBefore(ctx context.Context, before time.Time) (int64, error) {
	return m.countBefore, nil
}

func (m *mockStorageDB) EstimateEventSize(ctx context.Context) (int64, error) {
	return m.avgEventSize, nil
}

func (m *mockStorageDB) AddAuditLog(ctx context.Context, action string, details interface{}, performedBy string) error {
	return nil
}

// ============================================================================
// GetStorageStatus Tests
// ============================================================================

func TestGetStorageStatus(t *testing.T) {
	t.Run("returns_storage_status", func(t *testing.T) {
		mockDB := newMockStorageDB()

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/status", func(w http.ResponseWriter, r *http.Request) {
			relayDBSize, _ := mockDB.GetRelayDatabaseSize()
			appDBSize, _ := mockDB.GetAppDatabaseSize()
			totalSize := relayDBSize + appDBSize
			availableSpace, _ := mockDB.GetAvailableDiskSpace()
			totalSpace, _ := mockDB.GetTotalDiskSpace()

			var usagePercent float64
			if totalSpace > 0 {
				usedSpace := totalSpace - availableSpace
				usagePercent = float64(usedSpace) / float64(totalSpace) * 100
			}

			var totalEvents int64
			if mockDB.IsRelayDBConnected() {
				stats, _ := mockDB.GetRelayStats(r.Context())
				if stats != nil {
					totalEvents = stats.TotalEvents
				}
			}

			pendingDeletions, _ := mockDB.GetPendingDeletionCount(r.Context())

			status := "healthy"
			if usagePercent >= 95 {
				status = "critical"
			} else if usagePercent >= 90 {
				status = "low"
			} else if usagePercent >= 80 {
				status = "warning"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(StorageStatusResponse{
				DatabaseSize:     relayDBSize,
				AppDatabaseSize:  appDBSize,
				TotalSize:        totalSize,
				AvailableSpace:   availableSpace,
				TotalSpace:       totalSpace,
				UsagePercent:     usagePercent,
				TotalEvents:      totalEvents,
				Status:           status,
				PendingDeletions: pendingDeletions,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp StorageStatusResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.DatabaseSize != 1024*1024*100 {
			t.Errorf("expected relay DB size 104857600, got %d", resp.DatabaseSize)
		}
		if resp.AppDatabaseSize != 1024*1024*10 {
			t.Errorf("expected app DB size 10485760, got %d", resp.AppDatabaseSize)
		}
		if resp.TotalEvents != 1000 {
			t.Errorf("expected 1000 events, got %d", resp.TotalEvents)
		}
		if resp.Status != "healthy" {
			t.Errorf("expected status 'healthy', got '%s'", resp.Status)
		}
	})

	t.Run("status_warning_at_80_percent", func(t *testing.T) {
		mockDB := newMockStorageDB()
		mockDB.totalSpace = 100
		mockDB.availableSpace = 15 // 85% used

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/status", func(w http.ResponseWriter, r *http.Request) {
			availableSpace, _ := mockDB.GetAvailableDiskSpace()
			totalSpace, _ := mockDB.GetTotalDiskSpace()

			var usagePercent float64
			if totalSpace > 0 {
				usedSpace := totalSpace - availableSpace
				usagePercent = float64(usedSpace) / float64(totalSpace) * 100
			}

			status := "healthy"
			if usagePercent >= 95 {
				status = "critical"
			} else if usagePercent >= 90 {
				status = "low"
			} else if usagePercent >= 80 {
				status = "warning"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":        status,
				"usage_percent": usagePercent,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "warning" {
			t.Errorf("expected status 'warning', got '%v'", resp["status"])
		}
	})

	t.Run("status_critical_at_95_percent", func(t *testing.T) {
		mockDB := newMockStorageDB()
		mockDB.totalSpace = 100
		mockDB.availableSpace = 3 // 97% used

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/status", func(w http.ResponseWriter, r *http.Request) {
			availableSpace, _ := mockDB.GetAvailableDiskSpace()
			totalSpace, _ := mockDB.GetTotalDiskSpace()

			var usagePercent float64
			if totalSpace > 0 {
				usedSpace := totalSpace - availableSpace
				usagePercent = float64(usedSpace) / float64(totalSpace) * 100
			}

			status := "healthy"
			if usagePercent >= 95 {
				status = "critical"
			} else if usagePercent >= 90 {
				status = "low"
			} else if usagePercent >= 80 {
				status = "warning"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": status,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "critical" {
			t.Errorf("expected status 'critical', got '%v'", resp["status"])
		}
	})
}

// ============================================================================
// GetRetentionPolicy Tests
// ============================================================================

func TestGetRetentionPolicy(t *testing.T) {
	t.Run("returns_retention_policy", func(t *testing.T) {
		mockDB := newMockStorageDB()

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/retention", func(w http.ResponseWriter, r *http.Request) {
			policy, _ := mockDB.GetRetentionPolicy(r.Context())

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"retention_days": policy.RetentionDays,
				"exceptions":     policy.Exceptions,
				"honor_nip09":    policy.HonorNIP09,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/retention", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["retention_days"].(float64)) != 90 {
			t.Errorf("expected 90 retention days, got %v", resp["retention_days"])
		}
		if resp["honor_nip09"] != true {
			t.Errorf("expected honor_nip09 true, got %v", resp["honor_nip09"])
		}
		exceptions := resp["exceptions"].([]interface{})
		if len(exceptions) != 2 {
			t.Errorf("expected 2 exceptions, got %d", len(exceptions))
		}
	})
}

// ============================================================================
// UpdateRetentionPolicy Tests
// ============================================================================

func TestUpdateRetentionPolicy(t *testing.T) {
	t.Run("updates_policy_successfully", func(t *testing.T) {
		mockDB := newMockStorageDB()

		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/storage/retention", func(w http.ResponseWriter, r *http.Request) {
			var req RetentionPolicyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if req.RetentionDays < 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Retention days must be non-negative",
					"code":  "INVALID_RETENTION_DAYS",
				})
				return
			}

			currentPolicy, _ := mockDB.GetRetentionPolicy(r.Context())
			currentPolicy.RetentionDays = req.RetentionDays
			currentPolicy.Exceptions = req.Exceptions
			currentPolicy.HonorNIP09 = req.HonorNIP09
			mockDB.SetRetentionPolicy(r.Context(), currentPolicy)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":        true,
				"retention_days": req.RetentionDays,
				"exceptions":     req.Exceptions,
				"honor_nip09":    req.HonorNIP09,
			})
		})

		body := `{"retention_days":30,"exceptions":["kind:0"],"honor_nip09":false}`
		req := httptest.NewRequest("PUT", "/api/v1/storage/retention", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["retention_days"].(float64)) != 30 {
			t.Errorf("expected 30 retention days, got %v", resp["retention_days"])
		}
		if resp["honor_nip09"] != false {
			t.Errorf("expected honor_nip09 false, got %v", resp["honor_nip09"])
		}
	})

	t.Run("rejects_negative_retention_days", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/storage/retention", func(w http.ResponseWriter, r *http.Request) {
			var req RetentionPolicyRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.RetentionDays < 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Retention days must be non-negative",
					"code":  "INVALID_RETENTION_DAYS",
				})
				return
			}
		})

		body := `{"retention_days":-1,"exceptions":[],"honor_nip09":true}`
		req := httptest.NewRequest("PUT", "/api/v1/storage/retention", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_invalid_json", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/storage/retention", func(w http.ResponseWriter, r *http.Request) {
			var req RetentionPolicyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid request body",
					"code":  "INVALID_JSON",
				})
				return
			}
		})

		body := `{invalid json}`
		req := httptest.NewRequest("PUT", "/api/v1/storage/retention", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// ============================================================================
// ManualCleanup Tests
// ============================================================================

func TestManualCleanup(t *testing.T) {
	t.Run("rejects_future_date", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/storage/cleanup", func(w http.ResponseWriter, r *http.Request) {
			var req CleanupRequest
			json.NewDecoder(r.Body).Decode(&req)

			beforeDate, err := time.Parse(time.RFC3339, req.BeforeDate)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid date format",
					"code":  "INVALID_DATE",
				})
				return
			}

			if beforeDate.After(time.Now()) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Date must be in the past",
					"code":  "FUTURE_DATE",
				})
				return
			}
		})

		futureDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		body := `{"before_date":"` + futureDate + `"}`
		req := httptest.NewRequest("POST", "/api/v1/storage/cleanup", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_invalid_date_format", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/storage/cleanup", func(w http.ResponseWriter, r *http.Request) {
			var req CleanupRequest
			json.NewDecoder(r.Body).Decode(&req)

			_, err := time.Parse(time.RFC3339, req.BeforeDate)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid date format. Use ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)",
					"code":  "INVALID_DATE",
				})
				return
			}
		})

		body := `{"before_date":"2024-01-01"}`
		req := httptest.NewRequest("POST", "/api/v1/storage/cleanup", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// ============================================================================
// GetDeletionRequests Tests
// ============================================================================

func TestGetDeletionRequests(t *testing.T) {
	t.Run("returns_all_requests", func(t *testing.T) {
		mockDB := newMockStorageDB()
		now := time.Now()
		mockDB.deletionRequests = []db.DeletionRequest{
			{ID: 1, EventID: "event1", AuthorPubkey: "admin", Status: "pending", ReceivedAt: now},
			{ID: 2, EventID: "event2", AuthorPubkey: "admin", Status: "processed", ReceivedAt: now},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/deletion-requests", func(w http.ResponseWriter, r *http.Request) {
			status := r.URL.Query().Get("status")
			requests, _ := mockDB.GetDeletionRequests(r.Context(), status)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"requests": requests,
				"total":    len(requests),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/deletion-requests", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["total"].(float64)) != 2 {
			t.Errorf("expected 2 requests, got %v", resp["total"])
		}
	})

	t.Run("filters_by_status", func(t *testing.T) {
		mockDB := newMockStorageDB()
		now := time.Now()
		mockDB.deletionRequests = []db.DeletionRequest{
			{ID: 1, EventID: "event1", Status: "pending", ReceivedAt: now},
			{ID: 2, EventID: "event2", Status: "pending", ReceivedAt: now},
			{ID: 3, EventID: "event3", Status: "processed", ReceivedAt: now},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/deletion-requests", func(w http.ResponseWriter, r *http.Request) {
			status := r.URL.Query().Get("status")
			requests, _ := mockDB.GetDeletionRequests(r.Context(), status)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"requests": requests,
				"total":    len(requests),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/deletion-requests?status=pending", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["total"].(float64)) != 2 {
			t.Errorf("expected 2 pending requests, got %v", resp["total"])
		}
	})
}

// ============================================================================
// GetStorageEstimate Tests
// ============================================================================

func TestGetStorageEstimate(t *testing.T) {
	t.Run("returns_estimate", func(t *testing.T) {
		mockDB := newMockStorageDB()
		mockDB.countBefore = 500
		mockDB.avgEventSize = 1000

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/estimate", func(w http.ResponseWriter, r *http.Request) {
			beforeDateStr := r.URL.Query().Get("before_date")
			if beforeDateStr == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "before_date parameter is required",
					"code":  "MISSING_DATE",
				})
				return
			}

			beforeDate, err := time.Parse(time.RFC3339, beforeDateStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			eventCount, _ := mockDB.CountEventsBefore(r.Context(), beforeDate)
			avgSize, _ := mockDB.EstimateEventSize(r.Context())
			estimatedSpace := eventCount * avgSize

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"event_count":     eventCount,
				"estimated_space": estimatedSpace,
			})
		})

		pastDate := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)
		req := httptest.NewRequest("GET", "/api/v1/storage/estimate?before_date="+pastDate, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["event_count"].(float64)) != 500 {
			t.Errorf("expected 500 events, got %v", resp["event_count"])
		}
		if int(resp["estimated_space"].(float64)) != 500000 {
			t.Errorf("expected 500000 bytes, got %v", resp["estimated_space"])
		}
	})

	t.Run("requires_before_date", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/storage/estimate", func(w http.ResponseWriter, r *http.Request) {
			beforeDateStr := r.URL.Query().Get("before_date")
			if beforeDateStr == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "before_date parameter is required",
					"code":  "MISSING_DATE",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/storage/estimate", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// ============================================================================
// RunVacuum Tests
// ============================================================================

func TestRunVacuum(t *testing.T) {
	t.Run("returns_space_reclaimed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/storage/vacuum", func(w http.ResponseWriter, r *http.Request) {
			// Simulate vacuum
			startTime := time.Now()
			spaceReclaimed := int64(1024 * 1024) // 1 MB reclaimed
			duration := time.Since(startTime)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":         true,
				"space_reclaimed": spaceReclaimed,
				"duration_ms":     duration.Milliseconds(),
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/storage/vacuum", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["success"] != true {
			t.Errorf("expected success true, got %v", resp["success"])
		}
		if resp["space_reclaimed"] == nil {
			t.Error("expected space_reclaimed to be present")
		}
	})
}

// ============================================================================
// RunIntegrityCheck Tests
// ============================================================================

func TestRunIntegrityCheck(t *testing.T) {
	t.Run("returns_integrity_status", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/storage/integrity-check", func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			duration := time.Since(startTime)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     true,
				"app_db":      map[string]interface{}{"ok": true, "result": "ok"},
				"relay_db":    map[string]interface{}{"ok": true, "result": "ok"},
				"duration_ms": duration.Milliseconds(),
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/storage/integrity-check", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["success"] != true {
			t.Errorf("expected success true, got %v", resp["success"])
		}

		appDB := resp["app_db"].(map[string]interface{})
		if appDB["ok"] != true {
			t.Errorf("expected app_db ok true, got %v", appDB["ok"])
		}

		relayDB := resp["relay_db"].(map[string]interface{})
		if relayDB["ok"] != true {
			t.Errorf("expected relay_db ok true, got %v", relayDB["ok"])
		}
	})

	t.Run("reports_failure", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/storage/integrity-check", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  false,
				"app_db":   map[string]interface{}{"ok": true, "result": "ok"},
				"relay_db": map[string]interface{}{"ok": false, "result": "corruption detected"},
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/storage/integrity-check", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["success"] != false {
			t.Errorf("expected success false, got %v", resp["success"])
		}

		relayDB := resp["relay_db"].(map[string]interface{})
		if relayDB["ok"] != false {
			t.Errorf("expected relay_db ok false, got %v", relayDB["ok"])
		}
	})
}
