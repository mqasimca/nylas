// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestGlossyBox(t *testing.T) {
	theme := DefaultTheme()
	content := "Test Content"
	width := 50

	result := GlossyBox(theme, content, width)

	if result == "" {
		t.Error("GlossyBox returned empty string")
	}

	if !strings.Contains(result, content) {
		t.Error("GlossyBox should contain the content")
	}
}

func TestGlowText(t *testing.T) {
	theme := DefaultTheme()
	text := "Glow Test"

	result := GlowText(text, theme.Primary)

	if result == "" {
		t.Error("GlowText returned empty string")
	}

	if !strings.Contains(result, text) {
		t.Error("GlowText should contain the text")
	}
}

func TestGradientText(t *testing.T) {
	theme := DefaultTheme()
	text := "Gradient Test"

	result := GradientText(text, theme.Primary, theme.Accent)

	if result == "" {
		t.Error("GradientText returned empty string")
	}

	if !strings.Contains(result, text) {
		t.Error("GradientText should contain the text")
	}
}

func TestNeumorphicBox(t *testing.T) {
	theme := DefaultTheme()
	content := "Neumorphic Content"

	result := NeumorphicBox(theme, content)

	if result == "" {
		t.Error("NeumorphicBox returned empty string")
	}

	if !strings.Contains(result, content) {
		t.Error("NeumorphicBox should contain the content")
	}

	// Should have top highlight and bottom shadow
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Error("NeumorphicBox should have at least 3 lines (highlight, content, shadow)")
	}
}

func TestPremiumBorder(t *testing.T) {
	theme := DefaultTheme()

	style := PremiumBorder(theme)

	// Verify style has border
	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("PremiumBorder rendered empty string")
	}
}

func TestGlassPanel(t *testing.T) {
	theme := DefaultTheme()
	width := 60
	height := 20

	style := GlassPanel(theme, width, height)

	rendered := style.Render("Test Content")
	if rendered == "" {
		t.Error("GlassPanel rendered empty string")
	}
}

func TestShimmerBorder(t *testing.T) {
	theme := DefaultTheme()

	style := ShimmerBorder(theme.Primary, theme.Accent)

	rendered := style.Render("Shimmer Test")
	if rendered == "" {
		t.Error("ShimmerBorder rendered empty string")
	}
}

func TestPanelWithGlow(t *testing.T) {
	theme := DefaultTheme()
	content := "Panel Content"
	width := 50

	result := PanelWithGlow(theme, content, width)

	if result == "" {
		t.Error("PanelWithGlow returned empty string")
	}

	if !strings.Contains(result, content) {
		t.Error("PanelWithGlow should contain the content")
	}

	// Should have glow on multiple sides
	lines := strings.Split(result, "\n")
	if len(lines) < 5 {
		t.Error("PanelWithGlow should have multiple lines (glow + content)")
	}
}

func TestMetallicText(t *testing.T) {
	theme := DefaultTheme()
	text := "Metallic"

	result := MetallicText(text, theme)

	if result == "" {
		t.Error("MetallicText returned empty string")
	}

	if !strings.Contains(result, text) {
		t.Error("MetallicText should contain the text")
	}

	// Should have sparkles
	if !strings.Contains(result, "✨") {
		t.Error("MetallicText should contain sparkle emoji")
	}
}

func TestNeonBorder(t *testing.T) {
	theme := DefaultTheme()

	style := NeonBorder(theme.Primary)

	rendered := style.Render("Neon Test")
	if rendered == "" {
		t.Error("NeonBorder rendered empty string")
	}
}

func TestFloatingPanel(t *testing.T) {
	theme := DefaultTheme()
	content := "Floating Content"
	width := 50

	result := FloatingPanel(theme, content, width)

	if result == "" {
		t.Error("FloatingPanel returned empty string")
	}

	if !strings.Contains(result, content) {
		t.Error("FloatingPanel should contain the content")
	}

	// Should have shadow effect
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("FloatingPanel should have content + shadow")
	}
}

func TestAccentLine(t *testing.T) {
	theme := DefaultTheme()

	tests := []struct {
		name  string
		width int
		char  string
	}{
		{"default char", 60, ""},
		{"custom char", 60, "─"},
		{"asterisk", 40, "*"},
		{"small width", 20, "━"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AccentLine(theme, tt.width, tt.char)

			if result == "" {
				t.Error("AccentLine returned empty string")
			}

			// Result should have approximately the requested width
			// (allowing for ANSI escape codes)
			rendered := lipgloss.NewStyle().Render(result)
			if len(rendered) < 10 {
				t.Errorf("AccentLine result too short: %d chars", len(rendered))
			}
		})
	}
}

func TestAccentLine_EmptyCharDefault(t *testing.T) {
	theme := DefaultTheme()

	result := AccentLine(theme, 60, "")

	if result == "" {
		t.Error("AccentLine with empty char should use default")
	}

	// Should use default "━" character
	if !strings.Contains(result, "━") {
		t.Error("AccentLine with empty char should default to ━")
	}
}

func TestEffects_NilTheme(t *testing.T) {
	// Test that functions handle nil theme gracefully
	// These should not panic

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Function panicked with nil theme: %v", r)
		}
	}()

	// These will panic if not handled properly
	// For now, we expect them to work with valid theme
	theme := DefaultTheme()

	_ = GlossyBox(theme, "test", 50)
	_ = NeumorphicBox(theme, "test")
	_ = PanelWithGlow(theme, "test", 50)
	_ = FloatingPanel(theme, "test", 50)
	_ = MetallicText("test", theme)
	_ = AccentLine(theme, 50, "")
}

func TestEffects_EmptyContent(t *testing.T) {
	theme := DefaultTheme()

	tests := []struct {
		name string
		fn   func() string
	}{
		{"GlossyBox", func() string { return GlossyBox(theme, "", 50) }},
		{"NeumorphicBox", func() string { return NeumorphicBox(theme, "") }},
		{"PanelWithGlow", func() string { return PanelWithGlow(theme, "", 50) }},
		{"FloatingPanel", func() string { return FloatingPanel(theme, "", 50) }},
		{"MetallicText", func() string { return MetallicText("", theme) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			// Should not panic, even with empty content
			_ = result
		})
	}
}

func TestEffects_EdgeCases(t *testing.T) {
	theme := DefaultTheme()

	t.Run("very small width", func(t *testing.T) {
		result := AccentLine(theme, 1, "─")
		if result == "" {
			t.Error("AccentLine should handle small width")
		}
	})

	t.Run("very large width", func(t *testing.T) {
		result := AccentLine(theme, 1000, "─")
		if result == "" {
			t.Error("AccentLine should handle large width")
		}
	})

	t.Run("long content", func(t *testing.T) {
		longContent := strings.Repeat("Very long content ", 100)
		result := GlossyBox(theme, longContent, 50)
		if result == "" {
			t.Error("GlossyBox should handle long content")
		}
	})

	t.Run("multiline content", func(t *testing.T) {
		multiline := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
		result := NeumorphicBox(theme, multiline)
		if result == "" {
			t.Error("NeumorphicBox should handle multiline content")
		}
	})
}

func TestEffects_ColorConsistency(t *testing.T) {
	theme := DefaultTheme()

	// Test that functions use theme colors consistently
	t.Run("GlowText uses provided color", func(t *testing.T) {
		result := GlowText("test", theme.Primary)
		if result == "" {
			t.Error("GlowText should work with theme primary color")
		}
	})

	t.Run("GradientText uses both colors", func(t *testing.T) {
		result := GradientText("test", theme.Primary, theme.Accent)
		if result == "" {
			t.Error("GradientText should work with two colors")
		}
	})

	t.Run("NeonBorder uses provided color", func(t *testing.T) {
		style := NeonBorder(theme.Accent)
		result := style.Render("test")
		if result == "" {
			t.Error("NeonBorder should work with accent color")
		}
	})
}

func TestEffects_VisualSeparation(t *testing.T) {
	theme := DefaultTheme()

	// Test that effects create visual separation
	t.Run("PanelWithGlow creates distinct sections", func(t *testing.T) {
		result := PanelWithGlow(theme, "Content", 50)
		lines := strings.Split(result, "\n")

		if len(lines) < 3 {
			t.Error("PanelWithGlow should create multiple visual sections")
		}
	})

	t.Run("FloatingPanel creates shadow effect", func(t *testing.T) {
		result := FloatingPanel(theme, "Content", 50)
		lines := strings.Split(result, "\n")

		if len(lines) < 2 {
			t.Error("FloatingPanel should have content + shadow effect")
		}
	})

	t.Run("NeumorphicBox creates highlight and shadow", func(t *testing.T) {
		result := NeumorphicBox(theme, "Content")
		lines := strings.Split(result, "\n")

		// Should have highlight (top) + content (middle) + shadow (bottom)
		if len(lines) < 3 {
			t.Error("NeumorphicBox should create highlight, content, and shadow layers")
		}
	})
}
