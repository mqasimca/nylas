package templates

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestFileStore_CRUD(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "templates.json")
	store := NewFileStore(path)
	ctx := context.Background()

	t.Run("List empty store", func(t *testing.T) {
		templates, err := store.List(ctx, "")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(templates) != 0 {
			t.Errorf("expected 0 templates, got %d", len(templates))
		}
	})

	t.Run("Create template", func(t *testing.T) {
		tpl := &domain.EmailTemplate{
			Name:     "Welcome Email",
			Subject:  "Welcome {{name}}!",
			HTMLBody: "Hello {{name}}, welcome to {{company}}!",
			Category: "onboarding",
		}

		created, err := store.Create(ctx, tpl)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if created.ID == "" {
			t.Error("expected ID to be generated")
		}
		if created.Name != "Welcome Email" {
			t.Errorf("expected name 'Welcome Email', got %q", created.Name)
		}
		if len(created.Variables) != 2 {
			t.Errorf("expected 2 variables, got %d: %v", len(created.Variables), created.Variables)
		}
		if created.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("Get template", func(t *testing.T) {
		// First create one
		tpl := &domain.EmailTemplate{
			Name:     "Test Get",
			Subject:  "Test Subject",
			HTMLBody: "Test body",
		}
		created, err := store.Create(ctx, tpl)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := store.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if got.Name != "Test Get" {
			t.Errorf("expected name 'Test Get', got %q", got.Name)
		}
	})

	t.Run("Get nonexistent template", func(t *testing.T) {
		_, err := store.Get(ctx, "nonexistent")
		if err != ErrTemplateNotFound {
			t.Errorf("expected ErrTemplateNotFound, got %v", err)
		}
	})

	t.Run("Update template", func(t *testing.T) {
		tpl := &domain.EmailTemplate{
			Name:     "Original Name",
			Subject:  "Original Subject",
			HTMLBody: "Original body",
		}
		created, err := store.Create(ctx, tpl)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		created.Name = "Updated Name"
		created.Subject = "Updated {{topic}}"

		updated, err := store.Update(ctx, created)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		if updated.Name != "Updated Name" {
			t.Errorf("expected name 'Updated Name', got %q", updated.Name)
		}
		if len(updated.Variables) != 1 || updated.Variables[0] != "topic" {
			t.Errorf("expected variables [topic], got %v", updated.Variables)
		}
		if updated.UpdatedAt.Before(updated.CreatedAt) {
			t.Error("expected UpdatedAt to be after CreatedAt")
		}
	})

	t.Run("Delete template", func(t *testing.T) {
		tpl := &domain.EmailTemplate{
			Name:     "To Delete",
			Subject:  "Delete me",
			HTMLBody: "Gone soon",
		}
		created, err := store.Create(ctx, tpl)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.Delete(ctx, created.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = store.Get(ctx, created.ID)
		if err != ErrTemplateNotFound {
			t.Errorf("expected ErrTemplateNotFound after delete, got %v", err)
		}
	})

	t.Run("Delete nonexistent template", func(t *testing.T) {
		err := store.Delete(ctx, "nonexistent")
		if err != ErrTemplateNotFound {
			t.Errorf("expected ErrTemplateNotFound, got %v", err)
		}
	})
}

func TestFileStore_ListWithCategory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "templates.json")
	store := NewFileStore(path)
	ctx := context.Background()

	// Create templates with different categories
	templates := []domain.EmailTemplate{
		{Name: "Welcome", Subject: "Welcome", HTMLBody: "Hi", Category: "onboarding"},
		{Name: "Password Reset", Subject: "Reset", HTMLBody: "Reset your password", Category: "security"},
		{Name: "Intro", Subject: "Intro", HTMLBody: "Introduction", Category: "onboarding"},
	}

	for i := range templates {
		_, err := store.Create(ctx, &templates[i])
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	tests := []struct {
		name     string
		category string
		want     int
	}{
		{"all templates", "", 3},
		{"onboarding category", "onboarding", 2},
		{"security category", "security", 1},
		{"case insensitive", "ONBOARDING", 2},
		{"nonexistent category", "marketing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.List(ctx, tt.category)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}
			if len(got) != tt.want {
				t.Errorf("expected %d templates, got %d", tt.want, len(got))
			}
		})
	}
}

func TestFileStore_IncrementUsage(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "templates.json")
	store := NewFileStore(path)
	ctx := context.Background()

	tpl := &domain.EmailTemplate{
		Name:     "Usage Test",
		Subject:  "Test",
		HTMLBody: "Body",
	}

	created, err := store.Create(ctx, tpl)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.UsageCount != 0 {
		t.Errorf("expected initial usage count 0, got %d", created.UsageCount)
	}

	// Increment usage
	for i := 0; i < 3; i++ {
		err = store.IncrementUsage(ctx, created.ID)
		if err != nil {
			t.Fatalf("IncrementUsage failed: %v", err)
		}
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.UsageCount != 3 {
		t.Errorf("expected usage count 3, got %d", got.UsageCount)
	}
}

func TestFileStore_Path(t *testing.T) {
	path := "/custom/path/templates.json"
	store := NewFileStore(path)

	if got := store.Path(); got != path {
		t.Errorf("expected path %q, got %q", path, got)
	}
}

func TestDefaultPath(t *testing.T) {
	// Save and restore XDG_CONFIG_HOME
	original := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() { _ = os.Setenv("XDG_CONFIG_HOME", original) })

	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		_ = os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		path := DefaultPath()
		expected := "/custom/config/nylas/templates.json"
		if path != expected {
			t.Errorf("expected %q, got %q", expected, path)
		}
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		path := DefaultPath()
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "nylas", "templates.json")
		if path != expected {
			t.Errorf("expected %q, got %q", expected, path)
		}
	})
}

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name  string
		texts []string
		want  []string
		wantN int
	}{
		{
			name:  "single variable",
			texts: []string{"Hello {{name}}!"},
			want:  []string{"name"},
			wantN: 1,
		},
		{
			name:  "multiple variables",
			texts: []string{"Hello {{name}}, welcome to {{company}}!"},
			want:  []string{"name", "company"},
			wantN: 2,
		},
		{
			name:  "duplicate variables",
			texts: []string{"Hello {{name}}, {{name}} again!"},
			want:  []string{"name"},
			wantN: 1,
		},
		{
			name:  "variables across multiple texts",
			texts: []string{"Subject: {{topic}}", "Body: {{name}} discusses {{topic}}"},
			want:  []string{"topic", "name"},
			wantN: 2,
		},
		{
			name:  "no variables",
			texts: []string{"Hello world!"},
			want:  nil,
			wantN: 0,
		},
		{
			name:  "variables with whitespace",
			texts: []string{"Hello {{ name }}!"},
			want:  []string{"name"},
			wantN: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVariables(tt.texts...)
			if len(got) != tt.wantN {
				t.Errorf("expected %d variables, got %d: %v", tt.wantN, len(got), got)
			}
		})
	}
}

func TestExpandVariables(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		vars        map[string]string
		want        string
		wantMissing []string
	}{
		{
			name: "all variables present",
			text: "Hello {{name}}, welcome to {{company}}!",
			vars: map[string]string{"name": "John", "company": "Acme"},
			want: "Hello John, welcome to Acme!",
		},
		{
			name:        "missing variable",
			text:        "Hello {{name}}, from {{company}}!",
			vars:        map[string]string{"name": "John"},
			want:        "Hello John, from {{company}}!",
			wantMissing: []string{"company"},
		},
		{
			name: "no variables",
			text: "Hello world!",
			vars: map[string]string{"name": "John"},
			want: "Hello world!",
		},
		{
			name:        "all variables missing",
			text:        "Hello {{name}}!",
			vars:        map[string]string{},
			want:        "Hello {{name}}!",
			wantMissing: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, missing := ExpandVariables(tt.text, tt.vars)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
			if len(missing) != len(tt.wantMissing) {
				t.Errorf("expected missing %v, got %v", tt.wantMissing, missing)
			}
		})
	}
}
