//go:build !integration

package air

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTemplateVariables(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantCount int
		wantVars  []string
	}{
		{
			name:      "single variable",
			text:      "Hello {{name}}",
			wantCount: 1,
			wantVars:  []string{"name"},
		},
		{
			name:      "multiple variables",
			text:      "Hello {{name}}, welcome to {{company}}",
			wantCount: 2,
			wantVars:  []string{"name", "company"},
		},
		{
			name:      "duplicate variables",
			text:      "Hi {{name}}, {{name}} is great!",
			wantCount: 1,
			wantVars:  []string{"name"},
		},
		{
			name:      "no variables",
			text:      "Hello world, no variables here",
			wantCount: 0,
			wantVars:  []string{},
		},
		{
			name:      "empty string",
			text:      "",
			wantCount: 0,
			wantVars:  []string{},
		},
		{
			name:      "complex template",
			text:      "Dear {{recipient}},\n\nThis is {{sender}} from {{company}}.\n\n{{message}}\n\nBest,\n{{sender}}",
			wantCount: 4,
			wantVars:  []string{"recipient", "sender", "company", "message"},
		},
		{
			name:      "underscores in variable names",
			text:      "From {{start_date}} to {{end_date}}",
			wantCount: 2,
			wantVars:  []string{"start_date", "end_date"},
		},
		{
			name:      "mixed content",
			text:      "Meeting with {{name}} at {curly braces} and {{location}}",
			wantCount: 2,
			wantVars:  []string{"name", "location"},
		},
		{
			name:      "invalid placeholders ignored",
			text:      "Hello {{ name }} and {{valid}}",
			wantCount: 1,
			wantVars:  []string{"valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTemplateVariables(tt.text)

			assert.Len(t, result, tt.wantCount)
			for _, v := range tt.wantVars {
				assert.Contains(t, result, v)
			}
		})
	}
}

func TestDeduplicateStrings(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output []string
	}{
		{
			name:   "no duplicates",
			input:  []string{"a", "b", "c"},
			output: []string{"a", "b", "c"},
		},
		{
			name:   "with duplicates",
			input:  []string{"a", "b", "a", "c", "b"},
			output: []string{"a", "b", "c"},
		},
		{
			name:   "all duplicates",
			input:  []string{"x", "x", "x"},
			output: []string{"x"},
		},
		{
			name:   "empty slice",
			input:  []string{},
			output: []string{},
		},
		{
			name:   "single element",
			input:  []string{"only"},
			output: []string{"only"},
		},
		{
			name:   "preserves order",
			input:  []string{"z", "a", "m", "a", "z"},
			output: []string{"z", "a", "m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateStrings(tt.input)

			assert.Equal(t, tt.output, result)
		})
	}
}

func TestDefaultTemplates(t *testing.T) {
	templates := defaultTemplates()

	// Check that we have expected default templates
	assert.NotEmpty(t, templates)

	// Check for expected templates
	expectedIDs := []string{
		"default-thanks",
		"default-intro",
		"default-followup",
		"default-meeting",
		"default-ooo",
	}

	for _, id := range expectedIDs {
		found := false
		for _, tmpl := range templates {
			if tmpl.ID == id {
				found = true
				assert.NotEmpty(t, tmpl.Name)
				assert.NotEmpty(t, tmpl.Body)
				assert.NotEmpty(t, tmpl.Shortcut)
				assert.True(t, tmpl.CreatedAt > 0)
				assert.True(t, tmpl.UpdatedAt > 0)
				break
			}
		}
		assert.True(t, found, "Expected template with ID %s", id)
	}
}

func TestDefaultTemplates_Variables(t *testing.T) {
	templates := defaultTemplates()

	// Find specific templates and check their variables
	for _, tmpl := range templates {
		switch tmpl.ID {
		case "default-thanks":
			// Thanks template has no variables
			assert.Empty(t, tmpl.Variables)
		case "default-intro":
			// Intro template has name, my_name, company, purpose
			assert.Contains(t, tmpl.Variables, "name")
			assert.Contains(t, tmpl.Variables, "my_name")
			assert.Contains(t, tmpl.Variables, "company")
		case "default-followup":
			// Follow up template has name, topic
			assert.Contains(t, tmpl.Variables, "name")
			assert.Contains(t, tmpl.Variables, "topic")
		case "default-meeting":
			// Meeting template has name, topic, time1, time2, time3
			assert.Contains(t, tmpl.Variables, "name")
			assert.Contains(t, tmpl.Variables, "topic")
		case "default-ooo":
			// OOO template has start_date, end_date, backup_contact
			assert.Contains(t, tmpl.Variables, "start_date")
			assert.Contains(t, tmpl.Variables, "end_date")
		}
	}
}

func TestEmailTemplate_Fields(t *testing.T) {
	template := EmailTemplate{
		ID:         "test-id",
		Name:       "Test Template",
		Subject:    "Test Subject: {{topic}}",
		Body:       "Hello {{name}}, this is about {{topic}}.",
		Shortcut:   "/test",
		Variables:  []string{"name", "topic"},
		Category:   "test",
		UsageCount: 5,
		CreatedAt:  1704067200,
		UpdatedAt:  1704153600,
		Metadata:   map[string]string{"key": "value"},
	}

	assert.Equal(t, "test-id", template.ID)
	assert.Equal(t, "Test Template", template.Name)
	assert.Equal(t, "Test Subject: {{topic}}", template.Subject)
	assert.Equal(t, "Hello {{name}}, this is about {{topic}}.", template.Body)
	assert.Equal(t, "/test", template.Shortcut)
	assert.Equal(t, []string{"name", "topic"}, template.Variables)
	assert.Equal(t, "test", template.Category)
	assert.Equal(t, 5, template.UsageCount)
	assert.Equal(t, int64(1704067200), template.CreatedAt)
	assert.Equal(t, int64(1704153600), template.UpdatedAt)
	assert.Equal(t, "value", template.Metadata["key"])
}
