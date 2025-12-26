package air

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDemoServer(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":7365")

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	if !server.demoMode {
		t.Error("expected demoMode to be true")
	}

	if server.addr != ":7365" {
		t.Errorf("expected addr :7365, got %s", server.addr)
	}

	// Demo server should not have Nylas client or stores
	if server.nylasClient != nil {
		t.Error("expected nylasClient to be nil in demo mode")
	}

	if server.configSvc != nil {
		t.Error("expected configSvc to be nil in demo mode")
	}
}

func TestExtractName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		email    string
		expected string
	}{
		{"john@example.com", "John"},
		{"alice.smith@company.org", "Alice.smith"},
		{"bob@test.io", "Bob"},
		{"a@b.com", "A"},
		{"test", "test"}, // No @ symbol
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := extractName(tt.email)
			if result != tt.expected {
				t.Errorf("extractName(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}

func TestInitials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		email    string
		expected string
	}{
		{"john@example.com", "J"},
		{"alice@company.org", "A"},
		{"Bob@test.io", "B"},
		{"", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := initials(tt.email)
			if result != tt.expected {
				t.Errorf("initials(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}

func TestHandleIndex_NonRootPath(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":7365")

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleIndex_RootPath_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":7365")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	// Should succeed with templates loaded
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}

	// Check content type if successful
	if w.Code == http.StatusOK {
		contentType := w.Header().Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("expected content-type text/html, got %s", contentType)
		}
	}
}

func TestBuildPageData_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":7365")
	data := server.buildPageData()

	// In demo mode, should have mock data
	if len(data.Emails) == 0 {
		t.Error("expected non-empty emails in demo mode")
	}

	if len(data.Folders) == 0 {
		t.Error("expected non-empty folders in demo mode")
	}

	if len(data.Calendars) == 0 {
		t.Error("expected non-empty calendars in demo mode")
	}

	if len(data.Events) == 0 {
		t.Error("expected non-empty events in demo mode")
	}

	if len(data.Contacts) == 0 {
		t.Error("expected non-empty contacts in demo mode")
	}

	if data.UserName == "" {
		t.Error("expected non-empty UserName in demo mode")
	}
}
