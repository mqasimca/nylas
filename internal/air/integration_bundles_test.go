//go:build integration
// +build integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_GetBundles(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/bundles", nil)
	w := httptest.NewRecorder()

	server.handleGetBundles(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var bundles []*Bundle
	if err := json.NewDecoder(w.Body).Decode(&bundles); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify default bundles exist
	bundleIDs := make(map[string]bool)
	for _, b := range bundles {
		bundleIDs[b.ID] = true
	}

	expectedBundles := []string{"newsletters", "receipts", "social", "updates", "promotions"}
	for _, id := range expectedBundles {
		if !bundleIDs[id] {
			t.Errorf("expected bundle %q not found", id)
		}
	}
}

func TestIntegration_CategorizeEmailNewsletter(t *testing.T) {
	server := testServer(t)

	body := map[string]string{
		"from":    "newsletter@company.com",
		"subject": "Your weekly digest is here",
		"emailId": "test-email-1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/categorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleBundleCategorize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["bundleId"] != "newsletters" {
		t.Errorf("expected bundleId 'newsletters', got %q", result["bundleId"])
	}
}

func TestIntegration_CategorizeEmailReceipt(t *testing.T) {
	server := testServer(t)

	body := map[string]string{
		"from":    "noreply@amazon.com",
		"subject": "Your order confirmation #12345",
		"emailId": "test-email-2",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/categorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleBundleCategorize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["bundleId"] != "receipts" {
		t.Errorf("expected bundleId 'receipts', got %q", result["bundleId"])
	}
}

func TestIntegration_CategorizeEmailSocial(t *testing.T) {
	server := testServer(t)

	body := map[string]string{
		"from":    "notifications@twitter.com",
		"subject": "You have new followers",
		"emailId": "test-email-3",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/categorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleBundleCategorize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["bundleId"] != "social" {
		t.Errorf("expected bundleId 'social', got %q", result["bundleId"])
	}
}

func TestIntegration_CategorizeEmailPromotion(t *testing.T) {
	server := testServer(t)

	body := map[string]string{
		"from":    "deals@shop.com",
		"subject": "SALE: 50% off everything today!",
		"emailId": "test-email-4",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/categorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleBundleCategorize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["bundleId"] != "promotions" {
		t.Errorf("expected bundleId 'promotions', got %q", result["bundleId"])
	}
}

func TestIntegration_CategorizeEmailNoMatch(t *testing.T) {
	server := testServer(t)

	body := map[string]string{
		"from":    "friend@gmail.com",
		"subject": "Hey, how are you doing?",
		"emailId": "test-email-5",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/categorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleBundleCategorize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Personal email should not match any bundle
	if result["bundleId"] != "" {
		t.Errorf("expected no bundleId for personal email, got %q", result["bundleId"])
	}
}

func TestIntegration_GetBundleEmails(t *testing.T) {
	server := testServer(t)

	// First categorize an email
	body := map[string]string{
		"from":    "newsletter@test.com",
		"subject": "Weekly newsletter",
		"emailId": "bundle-test-email",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/categorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleBundleCategorize(w, req)

	// Then get emails in bundle
	req = httptest.NewRequest(http.MethodGet, "/api/bundles/emails?bundleId=newsletters", nil)
	w = httptest.NewRecorder()

	server.handleGetBundleEmails(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var emailIds []string
	if err := json.NewDecoder(w.Body).Decode(&emailIds); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should contain our categorized email
	found := false
	for _, id := range emailIds {
		if id == "bundle-test-email" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find categorized email in bundle")
	}
}

func TestIntegration_UpdateBundle(t *testing.T) {
	server := testServer(t)

	bundle := Bundle{
		ID:        "newsletters",
		Name:      "My Newsletters",
		Icon:      "📧",
		Collapsed: false,
	}
	bodyBytes, _ := json.Marshal(bundle)

	req := httptest.NewRequest(http.MethodPut, "/api/bundles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateBundle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify update persisted
	req = httptest.NewRequest(http.MethodGet, "/api/bundles", nil)
	w = httptest.NewRecorder()
	server.handleGetBundles(w, req)

	var bundles []*Bundle
	if err := json.NewDecoder(w.Body).Decode(&bundles); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for _, b := range bundles {
		if b.ID == "newsletters" {
			if b.Name != "My Newsletters" {
				t.Errorf("expected name 'My Newsletters', got %q", b.Name)
			}
			if b.Collapsed {
				t.Error("expected collapsed to be false")
			}
			break
		}
	}
}
