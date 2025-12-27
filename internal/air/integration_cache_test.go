//go:build integration
// +build integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// CACHE INTEGRATION TESTS
// ================================

func TestIntegration_CacheStatus(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cache/status", nil)
	w := httptest.NewRecorder()

	server.handleCacheStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Cache status: enabled=%v, online=%v, sync_interval=%d min",
		resp.Enabled, resp.Online, resp.SyncInterval)

	if len(resp.Accounts) > 0 {
		for _, acc := range resp.Accounts {
			t.Logf("  Account: %s (emails=%d, events=%d, size=%d bytes)",
				acc.Email, acc.EmailCount, acc.EventCount, acc.SizeBytes)
		}
	}
}

func TestIntegration_CacheSettings(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cache/settings", nil)
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheSettingsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Cache settings: enabled=%v, max_size=%dMB, ttl=%d days, theme=%s",
		resp.Enabled, resp.MaxSizeMB, resp.TTLDays, resp.Theme)

	// Verify settings have reasonable values
	if resp.MaxSizeMB < 50 {
		t.Errorf("MaxSizeMB should be at least 50, got %d", resp.MaxSizeMB)
	}
	if resp.TTLDays < 1 {
		t.Errorf("TTLDays should be at least 1, got %d", resp.TTLDays)
	}
}

func TestIntegration_CacheSearch_EmptyQuery(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cache/search?q=", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Empty query should return empty results
	if len(resp.Results) != 0 {
		t.Errorf("expected empty results for empty query, got %d", len(resp.Results))
	}
}

func TestIntegration_CacheSearch_WithQuery(t *testing.T) {
	server := testServer(t)

	// Search for a common term
	req := httptest.NewRequest(http.MethodGet, "/api/cache/search?q=test", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Cache search for 'test': found %d results", len(resp.Results))

	for i, r := range resp.Results {
		if i < 5 { // Log first 5 results
			t.Logf("  [%s] %s - %s", r.Type, r.Title, r.Subtitle)
		}
	}
}

// ================================
