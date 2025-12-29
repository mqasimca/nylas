package models

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (c *Compose) View() tea.View {
	var sb strings.Builder

	// Title
	title := "Compose New Message"
	switch c.mode {
	case ComposeModeReply:
		title = "Reply"
	case ComposeModeReplyAll:
		title = "Reply All"
	case ComposeModeForward:
		title = "Forward"
	case ComposeModeDraft:
		title = "Edit Draft"
	}

	sb.WriteString(c.theme.Title.Render(title))
	sb.WriteString("\n\n")

	// To field
	sb.WriteString(c.renderField("To:", c.toInput.View(), c.validationErrors["to"], c.focusIndex == 0))

	// Cc field (if shown)
	if c.showCc {
		sb.WriteString(c.renderField("Cc:", c.ccInput.View(), "", c.focusIndex == 1))
	}

	// Bcc field (if shown)
	if c.showBcc {
		sb.WriteString(c.renderField("Bcc:", c.bccInput.View(), "", c.focusIndex == 2))
	}

	// Subject field
	sb.WriteString(c.renderField("Subject:", c.subjectInput.View(), c.validationErrors["subject"], c.focusIndex == 3))

	// Body field
	sb.WriteString(c.renderBodyField())

	// Status line
	sb.WriteString("\n")
	sb.WriteString(c.renderStatusLine())

	// Help text
	sb.WriteString("\n")
	sb.WriteString(c.renderHelp())

	return tea.NewView(sb.String())
}

// renderField renders a form field
func (c *Compose) renderField(label, value, errMsg string, focused bool) string {
	var sb strings.Builder

	// Label
	labelStyle := c.theme.Help
	if focused {
		labelStyle = lipgloss.NewStyle().Foreground(c.theme.Primary).Bold(true)
	}
	sb.WriteString(labelStyle.Render(fmt.Sprintf("%-10s", label)))

	// Value
	sb.WriteString(value)

	// Error message
	if errMsg != "" {
		sb.WriteString("  ")
		sb.WriteString(c.theme.Error_.Render(errMsg))
	}

	sb.WriteString("\n")
	return sb.String()
}

// renderBodyField renders the body field
func (c *Compose) renderBodyField() string {
	var sb strings.Builder

	// Label
	labelStyle := c.theme.Help
	if c.focusIndex == 4 {
		labelStyle = lipgloss.NewStyle().Foreground(c.theme.Primary).Bold(true)
	}
	sb.WriteString(labelStyle.Render("Body:"))
	sb.WriteString("\n")

	// Body text area
	sb.WriteString(c.bodyInput.View())

	return sb.String()
}

// renderStatusLine renders the status line with save status
func (c *Compose) renderStatusLine() string {
	var parts []string

	// Send status
	if c.sending {
		primaryStyle := lipgloss.NewStyle().Foreground(c.theme.Primary)
		parts = append(parts, primaryStyle.Render("⏳ Sending..."))
	}

	// Save status
	saveStatus := c.renderSaveStatus()
	if saveStatus != "" {
		parts = append(parts, saveStatus)
	}

	return strings.Join(parts, "  ")
}

// renderSaveStatus renders the save status indicator
func (c *Compose) renderSaveStatus() string {
	primaryStyle := lipgloss.NewStyle().Foreground(c.theme.Primary)
	warningStyle := lipgloss.NewStyle().Foreground(c.theme.Warning)
	successStyle := lipgloss.NewStyle().Foreground(c.theme.Success)

	switch c.saveStatus {
	case SaveStatusSaving:
		return primaryStyle.Render("⏳ Saving draft...")
	case SaveStatusSaved:
		elapsed := time.Since(c.lastSaveTime)
		if c.isDirty {
			return warningStyle.Render("● Unsaved changes")
		}
		return successStyle.Render(fmt.Sprintf("✓ Draft saved %s", formatElapsed(elapsed)))
	case SaveStatusError:
		return c.theme.Error_.Render("✗ Save failed (Ctrl+S to retry)")
	case SaveStatusUnsaved:
		return warningStyle.Render("● Unsaved changes")
	default:
		return ""
	}
}

// formatElapsed formats an elapsed duration
func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	hours := int(d.Hours())
	if hours == 1 {
		return "1 hour ago"
	}
	return fmt.Sprintf("%d hours ago", hours)
}

// renderHelp renders the help text
func (c *Compose) renderHelp() string {
	helps := []string{
		"Tab: next field",
		"Ctrl+A: send",
		"Ctrl+S: save draft",
	}

	if !c.showCc {
		helps = append(helps, "Ctrl+T: show Cc")
	}
	if !c.showBcc {
		helps = append(helps, "Ctrl+B: show Bcc")
	}

	helps = append(helps, "Esc: cancel")

	return c.theme.Help.Render(strings.Join(helps, "  "))
}
