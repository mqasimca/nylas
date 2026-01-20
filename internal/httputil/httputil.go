// Package httputil provides common HTTP utilities shared across servers.
package httputil

import (
	"encoding/json"
	"io"
	"net/http"
)

// MaxRequestBodySize is the maximum allowed request body size (1MB).
// This prevents memory exhaustion attacks via large payloads.
const MaxRequestBodySize = 1 << 20

// LimitedBody wraps a request body with a size limit.
// Returns a ReadCloser that will return an error if the body exceeds maxBytes.
func LimitedBody(w http.ResponseWriter, r *http.Request, maxBytes int64) io.ReadCloser {
	return http.MaxBytesReader(w, r.Body, maxBytes)
}

// WriteJSON writes a JSON response with the given status code.
// It sets the Content-Type header to application/json.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// DecodeJSON decodes JSON from the request body into the target.
// It uses LimitedBody to prevent oversized payloads.
func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	return json.NewDecoder(LimitedBody(w, r, MaxRequestBodySize)).Decode(target)
}
