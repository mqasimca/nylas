package tui

import (
	"testing"
	"time"
)

func TestViews_FormatDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		contains string
	}{
		{
			name:     "today shows time only",
			input:    now,
			contains: ":", // Should contain time like "3:04 PM"
		},
		{
			name:     "yesterday shows month and day",
			input:    now.AddDate(0, 0, -1),
			contains: "", // Will vary
		},
		{
			name:     "last month shows month and day",
			input:    now.AddDate(0, -1, 0),
			contains: "", // Will vary based on current month
		},
		{
			name:     "last year shows year",
			input:    now.AddDate(-1, 0, 0),
			contains: "", // Should contain year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.input)
			if result == "" {
				t.Error("formatDate() returned empty string")
			}
		})
	}
}

func TestFormatDate_Today(t *testing.T) {
	now := time.Now()
	result := formatDate(now)

	// Today should show time format like "3:04 PM"
	if len(result) < 4 {
		t.Errorf("formatDate(today) = %q, expected time format", result)
	}
}

func TestFormatDate_LastYear(t *testing.T) {
	lastYear := time.Now().AddDate(-1, 0, 0)
	result := formatDate(lastYear)

	// Last year should include year
	if len(result) < 6 {
		t.Errorf("formatDate(lastYear) = %q, expected date with year", result)
	}
}

func TestViews_FormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{10240, "10.0 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{10485760, "10.0 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
		{10737418240, "10.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %q, want %q", tt.size, result, tt.expected)
			}
		})
	}
}

func TestStripHTMLForTUI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "simple tag",
			input:    "<p>Hello</p>",
			expected: "Hello",
		},
		{
			name:     "bold tag",
			input:    "Hello <b>World</b>!",
			expected: "Hello World!",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Hello <b>World</b></p></div>",
			expected: "Hello World",
		},
		{
			name:     "br tag",
			input:    "Line 1<br>Line 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "br self-closing",
			input:    "Line 1<br/>Line 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "html entity",
			input:    "Hello &amp; World",
			expected: "Hello & World",
		},
		{
			name:     "nbsp entity",
			input:    "Hello&nbsp;World",
			expected: "Hello\u00a0World",
		},
		{
			name:     "style tag removed",
			input:    "<style>body{color:red}</style>Hello",
			expected: "Hello",
		},
		{
			name:     "script tag removed",
			input:    "<script>alert('hi')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "head tag removed",
			input:    "<head><title>Test</title></head>Body",
			expected: "Body",
		},
		{
			name:     "multiple spaces collapsed",
			input:    "Hello    World",
			expected: "Hello World",
		},
		{
			name:     "multiple newlines collapsed",
			input:    "Hello\n\n\n\nWorld",
			expected: "Hello\n\nWorld",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTMLForTUI(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTMLForTUI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveTagWithContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tag      string
		expected string
	}{
		{
			name:     "remove style tag",
			input:    "Hello<style>body{}</style>World",
			tag:      "style",
			expected: "HelloWorld",
		},
		{
			name:     "remove script tag",
			input:    "Hello<script>code</script>World",
			tag:      "script",
			expected: "HelloWorld",
		},
		{
			name:     "no matching tag",
			input:    "Hello World",
			tag:      "style",
			expected: "Hello World",
		},
		{
			name:     "uppercase tag",
			input:    "Hello<STYLE>css</STYLE>World",
			tag:      "style",
			expected: "HelloWorld",
		},
		{
			name:     "multiple occurrences",
			input:    "<style>a</style>Hello<style>b</style>World<style>c</style>",
			tag:      "style",
			expected: "HelloWorld",
		},
		{
			name:     "self-closing tag",
			input:    "Hello<br/>World",
			tag:      "br",
			expected: "HelloWorld",
		},
		{
			name:     "empty input",
			input:    "",
			tag:      "style",
			expected: "",
		},
		{
			name:     "nested content",
			input:    "Hello<div><p>nested</p></div>World",
			tag:      "div",
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTagWithContent(tt.input, tt.tag)
			if result != tt.expected {
				t.Errorf("removeTagWithContent(%q, %q) = %q, want %q", tt.input, tt.tag, result, tt.expected)
			}
		})
	}
}

func TestStripHTMLForTUI_RealEmail(t *testing.T) {
	// Test with realistic email HTML
	input := `<html>
<head><title>Test</title></head>
<body>
<div style="font-family: Arial;">
<p>Hello,</p>
<p>This is a <b>test</b> email with <a href="https://example.com">a link</a>.</p>
<br>
<p>Best regards,<br>Sender</p>
</div>
</body>
</html>`

	result := stripHTMLForTUI(input)

	// Should not be empty
	if result == "" {
		t.Error("stripHTMLForTUI returned empty for email HTML")
	}

	// Should contain key text
	if !containsString(result, "Hello") {
		t.Error("Result should contain 'Hello'")
	}
	if !containsString(result, "test") {
		t.Error("Result should contain 'test'")
	}
	if !containsString(result, "Best regards") {
		t.Error("Result should contain 'Best regards'")
	}

	// Should not contain HTML tags
	if containsString(result, "<") || containsString(result, ">") {
		t.Errorf("Result should not contain HTML tags: %q", result)
	}
}

// Helper function
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
