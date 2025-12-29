package air

import (
	"regexp"
	"time"
)

// =============================================================================
// Email Templates Helper Functions
// =============================================================================

// extractTemplateVariables extracts {{variable}} placeholders from text.
func extractTemplateVariables(text string) []string {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(text, -1)

	seen := make(map[string]bool)
	vars := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			vars = append(vars, match[1])
			seen[match[1]] = true
		}
	}
	return vars
}

// deduplicateStrings removes duplicate strings while preserving order.
func deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// defaultTemplates returns built-in templates.
func defaultTemplates() []EmailTemplate {
	now := time.Now().Unix()
	return []EmailTemplate{
		{
			ID:        "default-thanks",
			Name:      "Thank You",
			Shortcut:  "/thanks",
			Body:      "Thank you for your email. I appreciate you reaching out and will get back to you shortly.\n\nBest regards",
			Category:  "closing",
			Variables: []string{},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "default-intro",
			Name:      "Introduction",
			Shortcut:  "/intro",
			Subject:   "Introduction: {{my_name}} from {{company}}",
			Body:      "Hi {{name}},\n\nI hope this email finds you well. My name is {{my_name}}, and I'm reaching out from {{company}}.\n\n{{purpose}}\n\nI'd love to schedule a brief call to discuss further. Would you have 15-20 minutes available this week?\n\nBest regards,\n{{my_name}}",
			Category:  "greeting",
			Variables: []string{"name", "my_name", "company", "purpose"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "default-followup",
			Name:      "Follow Up",
			Shortcut:  "/followup",
			Subject:   "Following up: {{topic}}",
			Body:      "Hi {{name}},\n\nI wanted to follow up on {{topic}}. Have you had a chance to review my previous message?\n\nPlease let me know if you have any questions or need additional information.\n\nBest regards",
			Category:  "follow-up",
			Variables: []string{"name", "topic"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "default-meeting",
			Name:      "Meeting Request",
			Shortcut:  "/meeting",
			Subject:   "Meeting Request: {{topic}}",
			Body:      "Hi {{name}},\n\nI'd like to schedule a meeting to discuss {{topic}}.\n\nWould any of the following times work for you?\n- {{time1}}\n- {{time2}}\n- {{time3}}\n\nPlease let me know what works best, or feel free to suggest an alternative time.\n\nBest regards",
			Category:  "greeting",
			Variables: []string{"name", "topic", "time1", "time2", "time3"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "default-ooo",
			Name:      "Out of Office",
			Shortcut:  "/ooo",
			Subject:   "Out of Office: {{start_date}} - {{end_date}}",
			Body:      "Hi,\n\nThank you for your email. I am currently out of the office from {{start_date}} to {{end_date}} with limited access to email.\n\nFor urgent matters, please contact {{backup_contact}}.\n\nI will respond to your email upon my return.\n\nBest regards",
			Category:  "auto-reply",
			Variables: []string{"start_date", "end_date", "backup_contact"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
