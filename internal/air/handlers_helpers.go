package air

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// defaultTimeout is the default API request timeout.
const defaultTimeout = 30 * time.Second

// withTimeout creates a context with the default timeout.
// Returns the context and a cancel function that must be deferred.
func (s *Server) withTimeout(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), defaultTimeout)
}

// requireConfig checks if the Nylas client is configured.
// Returns true if configured, false if not (error response already written).
// Callers should return immediately when this returns false.
func (s *Server) requireConfig(w http.ResponseWriter) bool {
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return false
	}
	return true
}

// parseJSONBody decodes a JSON request body into the provided destination.
// Returns true if successful, false if not (error response already written).
// Callers should return immediately when this returns false.
func parseJSONBody[T any](w http.ResponseWriter, r *http.Request, dest *T) bool {
	if err := json.NewDecoder(limitedBody(w, r)).Decode(dest); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return false
	}
	return true
}

// handleDemoMode returns the demo response if in demo mode.
// Returns true if demo mode is active (response already written), false otherwise.
// Callers should return immediately when this returns true.
func (s *Server) handleDemoMode(w http.ResponseWriter, data any) bool {
	if s.demoMode {
		writeJSON(w, http.StatusOK, data)
		return true
	}
	return false
}

// requireMethod checks if the request method matches the expected method.
// Returns true if method matches, false if not (error response already written).
// Callers should return immediately when this returns false.
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// withAuthGrant combines demo mode check, config check, and grant ID retrieval.
// Returns the grant ID if all checks pass, or empty string if any check fails
// (appropriate error response already written).
//
// Usage:
//
//	grantID := s.withAuthGrant(w, demoResponse)
//	if grantID == "" {
//	    return
//	}
func (s *Server) withAuthGrant(w http.ResponseWriter, demoResponse any) string {
	if demoResponse != nil && s.handleDemoMode(w, demoResponse) {
		return ""
	}
	if !s.requireConfig(w) {
		return ""
	}
	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return ""
	}
	return grantID
}
