// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// GlossyBox creates a glossy box effect with shadows and gradients
func GlossyBox(theme *Theme, content string, width int) string {
	// Create shadow effect (darker background)
	shadow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1a1a1a")).
		Render(strings.Repeat("▓", width+4))

	// Main content box with gradient border
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		BorderBackground(lipgloss.Color("#0a0a0a")).
		Background(lipgloss.Color("#0d0d0d")).
		Padding(1, 2).
		Width(width).
		Render(content)

	// Add top shine effect
	shine := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Faint(true).
		Render(strings.Repeat("▀", width+4))

	// Combine: shine, box, shadow
	return lipgloss.JoinVertical(
		lipgloss.Left,
		shine,
		box,
		shadow,
	)
}

// GlowText creates glowing text effect
func GlowText(text string, col color.Color) string {
	// Inner glow (bright)
	return lipgloss.NewStyle().
		Foreground(col).
		Bold(true).
		Render(text)
}

// GradientText creates gradient text effect (simulated with color progression)
func GradientText(text string, startColor, endColor color.Color) string {
	// For now, use bold + primary color for impact
	// Full gradient would require per-character coloring
	return lipgloss.NewStyle().
		Foreground(startColor).
		Bold(true).
		Render(text)
}

// NeumorphicBox creates a neumorphic/glass effect box
func NeumorphicBox(theme *Theme, content string) string {
	// Top highlight
	highlight := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Faint(true).
		Render("▔")

	// Main box with soft shadow border
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Primary).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		Background(theme.Background).
		Padding(1, 2).
		Render(content)

	// Bottom shadow
	shadow := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Faint(true).
		Render("▁")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		highlight,
		box,
		shadow,
	)
}

// PremiumBorder creates a premium double-border effect
func PremiumBorder(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(theme.Primary).
		BorderBackground(lipgloss.Color("#1a1a1a")).
		Background(lipgloss.Color("#0a0a0a")).
		Padding(1, 2)
}

// GlassPanel creates a glass morphism effect
func GlassPanel(theme *Theme, width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Background(theme.Background).
		Width(width).
		Height(height).
		Padding(1, 2)
}

// ShimmerBorder creates an animated-looking border effect
func ShimmerBorder(primaryColor, accentColor color.Color) lipgloss.Style {
	// Use double border for more "premium" feel
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2)
}

// PanelWithGlow creates a panel with outer glow effect
func PanelWithGlow(theme *Theme, content string, width int) string {
	// Outer glow lines (top and bottom)
	glowColor := theme.Primary
	topGlow := lipgloss.NewStyle().
		Foreground(glowColor).
		Faint(true).
		Width(width + 6).
		Align(lipgloss.Center).
		Render(strings.Repeat("▀", width+6))

	bottomGlow := lipgloss.NewStyle().
		Foreground(glowColor).
		Faint(true).
		Width(width + 6).
		Align(lipgloss.Center).
		Render(strings.Repeat("▄", width+6))

	// Main panel
	panel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Background(lipgloss.Color("#0d0d0d")).
		Padding(1, 2).
		Width(width).
		Render(content)

	// Add side glows
	leftGlow := lipgloss.NewStyle().
		Foreground(glowColor).
		Faint(true).
		Render("▌")

	rightGlow := lipgloss.NewStyle().
		Foreground(glowColor).
		Faint(true).
		Render("▐")

	panelWithSides := lipgloss.JoinHorizontal(
		lipgloss.Center,
		leftGlow,
		panel,
		rightGlow,
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		topGlow,
		panelWithSides,
		bottomGlow,
	)
}

// MetallicText creates metallic/chrome text effect
func MetallicText(text string, theme *Theme) string {
	// Simulate metallic by using bold + gradient-like colors
	return lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true).
		Italic(false).
		Render(fmt.Sprintf("✨ %s ✨", text))
}

// NeonBorder creates a neon-style border
func NeonBorder(col color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(col).
		Bold(true).
		Padding(1, 2)
}

// FloatingPanel creates a panel that appears to float with shadow
func FloatingPanel(theme *Theme, content string, width int) string {
	// Shadow base
	shadow := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Faint(true).
		Width(width + 2).
		Align(lipgloss.Center).
		Render(strings.Repeat("▁", width+2))

	// Main panel elevated
	panel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Background(theme.Background).
		Padding(1, 2).
		Width(width).
		Render(content)

	// Offset shadow to create depth
	return lipgloss.JoinVertical(
		lipgloss.Left,
		panel,
		"  "+shadow, // Offset for 3D effect
	)
}

// AccentLine creates an accented horizontal line with gradient effect
func AccentLine(theme *Theme, width int, char string) string {
	if char == "" {
		char = "━"
	}

	// Create line with gradient colors (simulated)
	left := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Faint(true).
		Render(strings.Repeat(char, width/3))

	middle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true).
		Render(strings.Repeat(char, width/3))

	right := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Faint(true).
		Render(strings.Repeat(char, width/3))

	return lipgloss.JoinHorizontal(lipgloss.Left, left, middle, right)
}
