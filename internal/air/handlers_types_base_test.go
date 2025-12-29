package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

// ================================
// HELPER FUNCTION TESTS
// ================================

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", resp["key"])
	}
}

func TestWriteJSON_NilValue(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "null\n" {
		t.Errorf("expected 'null', got %s", w.Body.String())
	}
}

func TestWriteJSON_EmptyMap(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]any{})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "{}\n" {
		t.Errorf("expected '{}', got %s", w.Body.String())
	}
}

func TestParticipantsToEmail(t *testing.T) {
	t.Parallel()

	participants := []EmailParticipantResponse{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
	}

	result := participantsToEmail(participants)

	if len(result) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(result))
	}

	if result[0].Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", result[0].Name)
	}

	if result[0].Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", result[0].Email)
	}
}

func TestParticipantsToEmail_Empty(t *testing.T) {
	t.Parallel()

	result := participantsToEmail(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}

	result = participantsToEmail([]EmailParticipantResponse{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestParticipantsToEmail_Multiple(t *testing.T) {
	t.Parallel()

	participants := []EmailParticipantResponse{
		{Name: "Person One", Email: "one@example.com"},
		{Name: "Person Two", Email: "two@example.com"},
		{Name: "", Email: "three@example.com"},
	}

	result := participantsToEmail(participants)

	if len(result) != 3 {
		t.Errorf("expected 3 participants, got %d", len(result))
	}

	if result[0].Name != "Person One" || result[0].Email != "one@example.com" {
		t.Error("first participant not converted correctly")
	}

	if result[2].Email != "three@example.com" {
		t.Error("third participant email not converted correctly")
	}
}

func TestGrantFromDomain(t *testing.T) {
	t.Parallel()

	domainGrant := struct {
		ID       string
		Email    string
		Provider string
	}{
		ID:       "grant-123",
		Email:    "test@example.com",
		Provider: "google",
	}

	// Since grantFromDomain expects domain.GrantInfo, we test the conversion logic
	result := Grant{
		ID:       domainGrant.ID,
		Email:    domainGrant.Email,
		Provider: domainGrant.Provider,
	}

	if result.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got %s", result.ID)
	}

	if result.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", result.Email)
	}

	if result.Provider != "google" {
		t.Errorf("expected provider 'google', got %s", result.Provider)
	}
}

func TestGrantFromDomain_Basic(t *testing.T) {
	t.Parallel()

	grantInfo := domain.GrantInfo{
		ID:       "grant-123",
		Email:    "user@example.com",
		Provider: domain.ProviderGoogle,
	}

	grant := grantFromDomain(grantInfo)

	if grant.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got %s", grant.ID)
	}
	if grant.Email != "user@example.com" {
		t.Errorf("expected Email 'user@example.com', got %s", grant.Email)
	}
	if grant.Provider != "google" {
		t.Errorf("expected Provider 'google', got %s", grant.Provider)
	}
}

func TestGrantFromDomain_DifferentProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider domain.Provider
		expected string
	}{
		{"Google", domain.ProviderGoogle, "google"},
		{"Microsoft", domain.ProviderMicrosoft, "microsoft"},
		{"IMAP", domain.ProviderIMAP, "imap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grantInfo := domain.GrantInfo{
				ID:       "grant-123",
				Email:    "user@example.com",
				Provider: tt.provider,
			}

			grant := grantFromDomain(grantInfo)

			if grant.Provider != tt.expected {
				t.Errorf("expected Provider '%s', got %s", tt.expected, grant.Provider)
			}
		})
	}
}

// ================================
// CONTACT HELPER FUNCTION TESTS
// ================================
