package air

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// gzipResponseWriter wraps http.ResponseWriter to support gzip compression.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing.
func (w *gzipResponseWriter) WriteHeader(status int) {
	w.statusCode = status
	w.ResponseWriter.WriteHeader(status)
}

// Write compresses the response body.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// CompressionMiddleware adds gzip compression to responses.
// This significantly reduces bandwidth and improves load times.
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Don't compress already compressed formats
		if strings.HasSuffix(r.URL.Path, ".gz") ||
			strings.HasSuffix(r.URL.Path, ".jpg") ||
			strings.HasSuffix(r.URL.Path, ".jpeg") ||
			strings.HasSuffix(r.URL.Path, ".png") ||
			strings.HasSuffix(r.URL.Path, ".gif") ||
			strings.HasSuffix(r.URL.Path, ".woff") ||
			strings.HasSuffix(r.URL.Path, ".woff2") {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer func() {
			_ = gz.Close() // Error is non-actionable in deferred context
		}()

		// Wrap response writer
		gzw := &gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Length will change after compression

		next.ServeHTTP(gzw, r)
	})
}

// CacheMiddleware adds appropriate cache headers for static assets.
// This reduces server load and improves perceived performance.
func CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Static assets - cache for 1 year
		if strings.HasPrefix(path, "/static/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".woff") ||
			strings.HasSuffix(path, ".woff2") ||
			strings.HasSuffix(path, ".svg") ||
			strings.HasSuffix(path, ".ico") {
			// Other static files - cache for 1 hour
			w.Header().Set("Cache-Control", "public, max-age=3600")
		} else if strings.HasPrefix(path, "/api/") {
			// API responses - no cache (always fresh)
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		} else {
			// HTML pages - cache for 5 minutes
			w.Header().Set("Cache-Control", "public, max-age=300")
		}

		next.ServeHTTP(w, r)
	})
}

// PerformanceMonitoringMiddleware tracks request timing and adds performance headers.
// This helps identify slow endpoints and enables browser performance monitoring.
func PerformanceMonitoringMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create custom response writer to capture status code and add timing after response
		srw := &timingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			start:          start,
		}

		// Process request
		next.ServeHTTP(srw, r)
	})
}

// timingResponseWriter wraps http.ResponseWriter to add performance timing.
type timingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	start         time.Time
	headerWritten bool
}

// WriteHeader captures the status code and adds Server-Timing header.
func (w *timingResponseWriter) WriteHeader(code int) {
	if !w.headerWritten {
		w.statusCode = code

		// Add Server-Timing header before writing
		duration := time.Since(w.start)
		w.ResponseWriter.Header().Set("Server-Timing",
			"total;dur="+formatDuration(duration))

		w.headerWritten = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures headers are written before body.
func (w *timingResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// formatDuration formats duration in milliseconds with 2 decimal places.
func formatDuration(d time.Duration) string {
	ms := float64(d.Nanoseconds()) / 1e6
	// Use strconv for accurate formatting
	formatted := strconv.FormatFloat(ms, 'f', 2, 64)
	// Remove trailing zeros and decimal point if not needed
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return formatted
}

// SecurityHeadersMiddleware adds security headers to all responses.
// This improves security posture and prevents common attacks.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (relaxed for local development)
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: https:; "+
				"font-src 'self' data:; "+
				"connect-src 'self' https://api.us.nylas.com https://api.eu.nylas.com;")

		next.ServeHTTP(w, r)
	})
}

// MethodOverrideMiddleware allows using X-HTTP-Method-Override header.
// This enables REST methods in environments that only support GET/POST.
func MethodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			override := r.Header.Get("X-HTTP-Method-Override")
			if override != "" {
				r.Method = strings.ToUpper(override)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers for local development.
// This allows the frontend to make API requests.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from same origin (localhost)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-HTTP-Method-Override")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
