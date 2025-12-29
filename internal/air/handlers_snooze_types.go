package air

// =============================================================================
// Snooze Types
// =============================================================================

// SnoozedEmail represents a snoozed email.
type SnoozedEmail struct {
	EmailID        string `json:"email_id"`
	SnoozeUntil    int64  `json:"snooze_until"` // Unix timestamp
	OriginalFolder string `json:"original_folder,omitempty"`
	CreatedAt      int64  `json:"created_at"`
}

// SnoozeRequest represents a request to snooze an email.
type SnoozeRequest struct {
	EmailID     string `json:"email_id"`
	SnoozeUntil int64  `json:"snooze_until,omitempty"` // Explicit Unix timestamp
	Duration    string `json:"duration,omitempty"`     // Natural language: "1h", "2d", "tomorrow 9am"
}

// SnoozeResponse represents a snooze operation response.
type SnoozeResponse struct {
	Success     bool   `json:"success"`
	EmailID     string `json:"email_id"`
	SnoozeUntil int64  `json:"snooze_until"`
	Message     string `json:"message,omitempty"`
}

// parseError represents a parsing error for duration strings.
type parseError struct {
	input string
}

func (e *parseError) Error() string {
	return "cannot parse duration: " + e.input
}
