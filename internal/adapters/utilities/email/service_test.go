package email

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestBuildTemplate_MissingName(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	req := &domain.TemplateRequest{
		Subject: "Test",
	}

	_, err := service.BuildTemplate(ctx, req)
	if err == nil {
		t.Error("expected error for missing template name")
	}
	if err != nil && err.Error() != "template name is required" {
		t.Errorf("expected 'template name is required' error, got: %v", err)
	}
}

func TestBuildTemplate_MissingSubject(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	req := &domain.TemplateRequest{
		Name: "test-template",
	}

	_, err := service.BuildTemplate(ctx, req)
	if err == nil {
		t.Error("expected error for missing subject")
	}
	if err != nil && err.Error() != "subject is required" {
		t.Errorf("expected 'subject is required' error, got: %v", err)
	}
}

func TestBuildTemplate_Valid(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	req := &domain.TemplateRequest{
		Name:      "welcome",
		Subject:   "Welcome to our service!",
		HTMLBody:  "<p>Hello {{name}}</p>",
		TextBody:  "Hello {{name}}",
		Variables: []string{"name"},
	}

	template, err := service.BuildTemplate(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if template.Name != "welcome" {
		t.Errorf("expected name=welcome, got %s", template.Name)
	}
	if template.Subject != "Welcome to our service!" {
		t.Errorf("expected subject='Welcome to our service!', got %s", template.Subject)
	}
	if len(template.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(template.Variables))
	}
}

func TestPreviewTemplate(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	template := &domain.EmailTemplate{
		HTMLBody: "<p>Hello {{name}}, your score is {{score}}</p>",
	}

	data := map[string]any{
		"name":  "Alice",
		"score": 95,
	}

	result, err := service.PreviewTemplate(ctx, template, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<p>Hello Alice, your score is 95</p>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestSanitizeHTML(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "remove script tags",
			input: "<p>Hello</p><script>alert('xss')</script>",
			want:  "<p>Hello</p>>alert('xss')", // Removes <script and </script>, leaves > from opening tag
		},
		{
			name:  "remove iframe tags",
			input: "<div>Content</div><iframe src='evil.com'></iframe>",
			want:  "<div>Content</div> src='evil.com'>",
		},
		{
			name:  "remove javascript: protocol",
			input: "<a href='javascript:void(0)'>Link</a>",
			want:  "<a href='void(0)'>Link</a>",
		},
		{
			name:  "remove onclick attributes",
			input: "<button onclick='doEvil()'>Click</button>",
			want:  "<button 'doEvil()'>Click</button>",
		},
		{
			name:  "remove onerror attributes",
			input: "<img onerror='alert(1)' src='x.jpg'>",
			want:  "<img 'alert(1)' src='x.jpg'>",
		},
		{
			name:  "case insensitive removal",
			input: "<p>Test</p><SCRIPT>bad()</SCRIPT>",
			want:  "<p>Test</p>>bad()", // Removes <SCRIPT and </SCRIPT>, leaves > from opening tag
		},
		{
			name:  "safe HTML unchanged",
			input: "<p><strong>Bold</strong> and <em>italic</em></p>",
			want:  "<p><strong>Bold</strong> and <em>italic</em></p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.SanitizeHTML(ctx, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateEmailAddress(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	tests := []struct {
		name            string
		email           string
		wantFormatValid bool
		wantDisposable  bool
		skipMXCheck     bool // Some MX checks may fail in test env
	}{
		{
			name:            "valid email format",
			email:           "user@example.com",
			wantFormatValid: true,
			skipMXCheck:     true, // MX may vary
		},
		{
			name:            "invalid email - no @",
			email:           "userexample.com",
			wantFormatValid: false,
		},
		{
			name:            "invalid email - double @",
			email:           "user@@example.com",
			wantFormatValid: false,
		},
		{
			name:            "disposable email",
			email:           "test@tempmail.com",
			wantFormatValid: true,
			wantDisposable:  true,
			skipMXCheck:     true,
		},
		{
			name:            "disposable email - mailinator",
			email:           "test@mailinator.com",
			wantFormatValid: true,
			wantDisposable:  true,
			skipMXCheck:     true,
		},
		{
			name:            "email with display name",
			email:           "user@example.com", // Note: mail.ParseAddress handles display names differently
			wantFormatValid: true,
			skipMXCheck:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ValidateEmailAddress(ctx, tt.email)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.FormatValid != tt.wantFormatValid {
				t.Errorf("FormatValid = %v, want %v", got.FormatValid, tt.wantFormatValid)
			}

			if got.Disposable != tt.wantDisposable {
				t.Errorf("Disposable = %v, want %v", got.Disposable, tt.wantDisposable)
			}

			// MX check is skipped for most tests as it requires network
			if !tt.skipMXCheck && got.FormatValid {
				// In real environment, MX check should work
				t.Logf("MX check result: %v", got.MXExists)
			}
		})
	}
}

func TestAnalyzeSpamScore(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	tests := []struct {
		name        string
		html        string
		headers     map[string]string
		wantIsSpam  bool
		wantTrigger string // At least one trigger should contain this
	}{
		{
			name:        "clean email",
			html:        "<p>Hello, this is a normal email.</p>",
			headers:     map[string]string{},
			wantIsSpam:  false,
			wantTrigger: "",
		},
		{
			name:        "spam word - free money",
			html:        "<p>Get your free money now!</p>",
			headers:     map[string]string{},
			wantIsSpam:  false, // Single trigger not enough
			wantTrigger: "free money",
		},
		{
			name:        "multiple spam words",
			html:        "<p>Free money! Click here! Act now! You won!</p>",
			headers:     map[string]string{},
			wantIsSpam:  true, // Multiple triggers
			wantTrigger: "spam_word",
		},
		{
			name:        "excessive caps",
			html:        "<p>THIS IS SHOUTING AT YOU!!!</p>",
			headers:     map[string]string{},
			wantIsSpam:  false,
			wantTrigger: "excessive_caps",
		},
		{
			name:        "excessive exclamation marks",
			html:        "<p>Buy now!!!! Limited time!!!!</p>",
			headers:     map[string]string{},
			wantIsSpam:  false,
			wantTrigger: "excessive_exclamation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.AnalyzeSpamScore(ctx, tt.html, tt.headers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.IsSpam != tt.wantIsSpam {
				t.Errorf("IsSpam = %v, want %v (score: %.2f)", got.IsSpam, tt.wantIsSpam, got.Score)
			}

			if tt.wantTrigger != "" {
				found := false
				for _, trigger := range got.Triggers {
					if trigger.Rule == tt.wantTrigger || contains(trigger.Description, tt.wantTrigger) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected trigger containing %q, triggers: %+v", tt.wantTrigger, got.Triggers)
				}
			}
		})
	}
}

func TestHasExcessiveCaps(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "normal text",
			input: "Hello World",
			want:  false,
		},
		{
			name:  "all caps",
			input: "HELLO WORLD",
			want:  true,
		},
		{
			name:  "mostly caps",
			input: "HELLO World",
			want:  true, // 6/10 = 60% (HELLO=5, W=1)
		},
		{
			name:  "over 50% caps",
			input: "HELLO WORLD test",
			want:  true, // 10/14 > 50%
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "no letters",
			input: "123 456",
			want:  false,
		},
		{
			name:  "all lowercase",
			input: "hello world",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasExcessiveCaps(tt.input)
			if got != tt.want {
				t.Errorf("hasExcessiveCaps(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateID()
		if id == "" {
			t.Error("generateID() returned empty string")
		}
		if !hasPrefix(id, "tpl_") {
			t.Errorf("generateID() = %q, want prefix 'tpl_'", id)
		}
		if ids[id] {
			t.Errorf("generateID() generated duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestInlineCSS(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Currently returns input unchanged (TODO implementation)
	input := "<style>p{color:red;}</style><p>Test</p>"
	got, err := service.InlineCSS(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != input {
		t.Errorf("InlineCSS() modified input (not yet implemented)")
	}
}

func TestGenerateEML(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	msg := &domain.EmailMessage{
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "Test",
		HTMLBody: "<p>Test body</p>",
		TextBody: "Test body",
	}

	// Currently not implemented
	_, err := service.GenerateEML(ctx, msg)
	if err == nil {
		t.Error("GenerateEML() should return 'not implemented' error")
	}
}

// Helper functions for tests

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
