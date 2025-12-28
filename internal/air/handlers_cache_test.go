package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ================================
// CACHE HANDLER TESTS
// ================================

func TestHandleCacheStatus_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/status", nil)
	w := httptest.NewRecorder()

	server.handleCacheStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode should return mock status
	if !resp.Enabled {
		t.Error("expected Enabled to be true in demo mode")
	}

	if !resp.Online {
		t.Error("expected Online to be true in demo mode")
	}
}

func TestHandleCacheStatus_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/status", nil)
	w := httptest.NewRecorder()

	server.handleCacheStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSync_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/sync", nil)
	w := httptest.NewRecorder()

	server.handleCacheSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSyncResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true in demo mode")
	}

	if resp.Message == "" {
		t.Error("expected Message to be set")
	}
}

func TestHandleCacheSync_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/sync", nil)
	w := httptest.NewRecorder()

	server.handleCacheSync(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheClear_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/clear", nil)
	w := httptest.NewRecorder()

	server.handleCacheClear(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSyncResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true in demo mode")
	}
}

func TestHandleCacheClear_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/clear", nil)
	w := httptest.NewRecorder()

	server.handleCacheClear(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSearch_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/search?q=test", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode returns mock search results
	if resp.Query != "test" {
		t.Errorf("expected Query 'test', got %s", resp.Query)
	}

	if len(resp.Results) == 0 {
		t.Error("expected non-empty search results in demo mode")
	}
}

func TestHandleCacheSearch_EmptyQuery(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/search", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
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

func TestHandleCacheSearch_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/search", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSettings_GET_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/settings", nil)
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSettingsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode should return default settings
	if !resp.Enabled {
		t.Error("expected Enabled to be true")
	}

	if resp.MaxSizeMB == 0 {
		t.Error("expected MaxSizeMB to be set")
	}

	if resp.Theme == "" {
		t.Error("expected Theme to be set")
	}
}

func TestHandleCacheSettings_PUT_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{
		"cache_enabled": true,
		"cache_max_size_mb": 1000,
		"theme": "light"
	}`
	req := httptest.NewRequest(http.MethodPut, "/api/cache/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		t.Error("expected success to be true")
	}
}

func TestHandleCacheSettings_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/cache/settings", nil)
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSync_WithEmail(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/sync?email=test@example.com", nil)
	w := httptest.NewRecorder()

	server.handleCacheSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCacheClear_WithEmail(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/clear?email=test@example.com", nil)
	w := httptest.NewRecorder()

	server.handleCacheClear(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
