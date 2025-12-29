package air

import (
	"strings"
	"testing"
)

func TestEmailBodyCSS_HasLightBackground(t *testing.T) {
	t.Parallel()

	// Read the preview.css file from embedded files
	cssContent, err := staticFiles.ReadFile("static/css/preview.css")
	if err != nil {
		t.Fatalf("failed to read preview.css: %v", err)
	}

	css := string(cssContent)

	// Verify email iframe container has white/light background for readability
	tests := []struct {
		name     string
		contains string
		reason   string
	}{
		{
			"email iframe container has light background",
			"background: #ffffff",
			"HTML emails have inline styles for light backgrounds - need white bg for readability",
		},
		{
			"email body selector exists",
			".email-detail-body",
			"Email body styling must be defined",
		},
		{
			"email iframe container selector exists",
			".email-iframe-container",
			"Email iframe container styling must be defined for sandboxed email rendering",
		},
		{
			"email iframe styling exists",
			".email-body-iframe",
			"Sandboxed iframe styling must be defined for security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(css, tt.contains) {
				t.Errorf("preview.css missing required style: %s\nReason: %s", tt.contains, tt.reason)
			}
		})
	}
}

// ================================
// RESPONSE CONVERTER TESTS
// ================================
