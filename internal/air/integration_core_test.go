//go:build integration
// +build integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ================================
// CONFIG INTEGRATION TESTS
// ================================

func TestIntegration_ConfigStatus(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ConfigStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Configured {
		t.Error("expected Configured to be true")
	}

	if !resp.HasAPIKey {
		t.Error("expected HasAPIKey to be true")
	}

	// Note: GrantCount and DefaultGrant may be empty if ConfigService
	// doesn't have access to the grant store. This is not a failure,
	// just log the values for debugging.
	t.Logf("Config: region=%s, grants=%d, default=%s", resp.Region, resp.GrantCount, resp.DefaultGrant)
}

// ================================
// GRANTS INTEGRATION TESTS
// ================================

func TestIntegration_ListGrants(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp GrantsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Grants) == 0 {
		t.Error("expected at least one grant")
	}

	if resp.DefaultGrant == "" {
		t.Error("expected DefaultGrant to be set")
	}

	// Find the Google grant
	hasGoogle := false
	for _, g := range resp.Grants {
		if g.Provider == "google" {
			hasGoogle = true
			t.Logf("Found Google grant: %s (%s)", g.Email, g.ID)
		}
	}

	if !hasGoogle {
		t.Error("expected at least one Google grant")
	}
}

// ================================
// FOLDERS INTEGRATION TESTS
// ================================

func TestIntegration_ListFolders(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/folders", nil)
	w := httptest.NewRecorder()

	server.handleListFolders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp FoldersResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Folders) == 0 {
		t.Error("expected at least one folder")
	}

	// Check for standard Gmail folders/labels
	folderNames := make(map[string]bool)
	for _, f := range resp.Folders {
		folderNames[f.Name] = true
		if f.SystemFolder != "" {
			t.Logf("Folder: %s (system: %s, unread: %d)", f.Name, f.SystemFolder, f.UnreadCount)
		}
	}

	// Gmail should have INBOX
	if !folderNames["INBOX"] && !folderNames["Inbox"] {
		t.Log("Warning: INBOX folder not found (may have different name)")
	}
}

// ================================
// INDEX PAGE INTEGRATION TESTS
// ================================

func TestIntegration_IndexPage(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	// Check for basic HTML structure
	if !containsAny(body, "<html", "<HTML") {
		t.Error("expected HTML content")
	}

	// Check for Air UI elements
	if !containsAny(body, "Nylas Air", "email-list", "calendar-view") {
		t.Error("expected Air UI elements")
	}

	// Should have JavaScript initialization
	if !containsAny(body, "skeleton", "loading") {
		t.Log("Warning: No skeleton/loading indicators found")
	}

	t.Logf("Index page rendered successfully (%d bytes)", len(body))
}

func TestIntegration_BuildPageData(t *testing.T) {
	server := testServer(t)

	data := server.buildPageData()

	// Should have real user info
	if data.UserEmail == "" {
		t.Error("expected UserEmail to be set")
	}

	if data.DefaultGrantID == "" {
		t.Error("expected DefaultGrantID to be set")
	}

	if data.Provider != "google" {
		t.Errorf("expected Provider 'google', got %s", data.Provider)
	}

	// In non-demo mode, mock data should be cleared
	if len(data.Emails) > 0 {
		t.Error("expected Emails to be empty (loaded via JS)")
	}

	if len(data.Events) > 0 {
		t.Error("expected Events to be empty (loaded via JS)")
	}

	t.Logf("Page data: user=%s, provider=%s, grants=%d",
		data.UserEmail, data.Provider, len(data.Grants))
}
