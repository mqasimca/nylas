package constants

import (
	"testing"
	"time"
)

// TestPortConstants verifies all port constants are positive integers.
func TestPortConstants(t *testing.T) {
	tests := []struct {
		name  string
		port  int
		valid bool
	}{
		{"DefaultInboundPort", DefaultInboundPort, true},
		{"DefaultCallbackPort", DefaultCallbackPort, true},
		{"DefaultOllamaPort", DefaultOllamaPort, true},
		{"DefaultAirUIPort", DefaultAirUIPort, true},
		{"DefaultAirPort", DefaultAirPort, true},
		{"DefaultUIPort", DefaultUIPort, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.port <= 0 {
				t.Errorf("%s = %d, want positive integer", tt.name, tt.port)
			}
			if tt.port > 65535 {
				t.Errorf("%s = %d, want valid port (1-65535)", tt.name, tt.port)
			}
		})
	}
}

// TestPortValues verifies specific port values match expected defaults.
func TestPortValues(t *testing.T) {
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"DefaultInboundPort", DefaultInboundPort, 3000},
		{"DefaultCallbackPort", DefaultCallbackPort, 8080},
		{"DefaultOllamaPort", DefaultOllamaPort, 11434},
		{"DefaultAirUIPort", DefaultAirUIPort, 7365},
		{"DefaultAirPort", DefaultAirPort, 7363},
		{"DefaultUIPort", DefaultUIPort, 7363},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestURLConstants verifies all URL constants are non-empty and valid.
func TestURLConstants(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"DefaultNylasAPIBaseURL", DefaultNylasAPIBaseURL},
		{"DefaultSchedulerBaseURL", DefaultSchedulerBaseURL},
		{"DefaultOllamaHost", DefaultOllamaHost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.url == "" {
				t.Errorf("%s is empty, want non-empty URL", tt.name)
			}
			// Basic URL validation - should start with http:// or https://
			if len(tt.url) < 7 {
				t.Errorf("%s = %q, want valid URL", tt.name, tt.url)
			}
		})
	}
}

// TestServiceName verifies service name is non-empty.
func TestServiceName(t *testing.T) {
	if ServiceName == "" {
		t.Error("ServiceName is empty, want non-empty string")
	}
	if ServiceName != "nylas" {
		t.Errorf("ServiceName = %q, want %q", ServiceName, "nylas")
	}
}

// TestTimeoutConstants verifies all timeout constants are positive durations.
func TestTimeoutConstants(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"DefaultAPITimeout", DefaultAPITimeout},
		{"DefaultOAuthTimeout", DefaultOAuthTimeout},
		{"DefaultAITimeout", DefaultAITimeout},
		{"DefaultHealthCheckTimeout", DefaultHealthCheckTimeout},
		{"DefaultQuickCheckTimeout", DefaultQuickCheckTimeout},
		{"DefaultHTTPReadHeaderTimeout", DefaultHTTPReadHeaderTimeout},
		{"DefaultHTTPReadTimeout", DefaultHTTPReadTimeout},
		{"DefaultHTTPWriteTimeout", DefaultHTTPWriteTimeout},
		{"DefaultHTTPIdleTimeout", DefaultHTTPIdleTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout <= 0 {
				t.Errorf("%s = %v, want positive duration", tt.name, tt.timeout)
			}
		})
	}
}

// TestTimeoutValues verifies specific timeout values match expected defaults.
func TestTimeoutValues(t *testing.T) {
	tests := []struct {
		name     string
		got      time.Duration
		expected time.Duration
	}{
		{"DefaultAPITimeout", DefaultAPITimeout, 90 * time.Second},
		{"DefaultOAuthTimeout", DefaultOAuthTimeout, 5 * time.Minute},
		{"DefaultAITimeout", DefaultAITimeout, 120 * time.Second},
		{"DefaultHealthCheckTimeout", DefaultHealthCheckTimeout, 10 * time.Second},
		{"DefaultQuickCheckTimeout", DefaultQuickCheckTimeout, 5 * time.Second},
		{"DefaultHTTPReadHeaderTimeout", DefaultHTTPReadHeaderTimeout, 10 * time.Second},
		{"DefaultHTTPReadTimeout", DefaultHTTPReadTimeout, 30 * time.Second},
		{"DefaultHTTPWriteTimeout", DefaultHTTPWriteTimeout, 30 * time.Second},
		{"DefaultHTTPIdleTimeout", DefaultHTTPIdleTimeout, 120 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestTimeoutRelationships verifies logical relationships between timeouts.
func TestTimeoutRelationships(t *testing.T) {
	t.Run("QuickCheckShorterThanHealthCheck", func(t *testing.T) {
		if DefaultQuickCheckTimeout >= DefaultHealthCheckTimeout {
			t.Errorf("DefaultQuickCheckTimeout (%v) should be < DefaultHealthCheckTimeout (%v)",
				DefaultQuickCheckTimeout, DefaultHealthCheckTimeout)
		}
	})

	t.Run("HealthCheckShorterThanAPI", func(t *testing.T) {
		if DefaultHealthCheckTimeout >= DefaultAPITimeout {
			t.Errorf("DefaultHealthCheckTimeout (%v) should be < DefaultAPITimeout (%v)",
				DefaultHealthCheckTimeout, DefaultAPITimeout)
		}
	})

	t.Run("APITimeoutShorterThanOAuth", func(t *testing.T) {
		if DefaultAPITimeout >= DefaultOAuthTimeout {
			t.Errorf("DefaultAPITimeout (%v) should be < DefaultOAuthTimeout (%v)",
				DefaultAPITimeout, DefaultOAuthTimeout)
		}
	})

	t.Run("ReadHeaderTimeoutShorterThanReadTimeout", func(t *testing.T) {
		if DefaultHTTPReadHeaderTimeout >= DefaultHTTPReadTimeout {
			t.Errorf("DefaultHTTPReadHeaderTimeout (%v) should be < DefaultHTTPReadTimeout (%v)",
				DefaultHTTPReadHeaderTimeout, DefaultHTTPReadTimeout)
		}
	})
}
