package air

// =============================================================================
// Send Later / Scheduled Send Types
// =============================================================================

// ScheduledSendRequest represents a request to schedule an email.
type ScheduledSendRequest struct {
	To            []EmailParticipantResponse `json:"to"`
	Cc            []EmailParticipantResponse `json:"cc,omitempty"`
	Bcc           []EmailParticipantResponse `json:"bcc,omitempty"`
	Subject       string                     `json:"subject"`
	Body          string                     `json:"body"`
	SendAt        int64                      `json:"send_at,omitempty"`         // Unix timestamp
	SendAtNatural string                     `json:"send_at_natural,omitempty"` // Natural language
}

// ScheduledSendResponse represents a scheduled send response.
type ScheduledSendResponse struct {
	Success    bool   `json:"success"`
	ScheduleID string `json:"schedule_id,omitempty"`
	SendAt     int64  `json:"send_at"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// =============================================================================
// Undo Send Types
// =============================================================================

// UndoSendConfig holds undo send configuration.
type UndoSendConfig struct {
	Enabled        bool `json:"enabled"`
	GracePeriodSec int  `json:"grace_period_sec"` // Default: 10 seconds
}

// PendingSend represents a message in the undo grace period.
type PendingSend struct {
	ID        string                     `json:"id"`
	To        []EmailParticipantResponse `json:"to"`
	Cc        []EmailParticipantResponse `json:"cc,omitempty"`
	Bcc       []EmailParticipantResponse `json:"bcc,omitempty"`
	Subject   string                     `json:"subject"`
	Body      string                     `json:"body"`
	CreatedAt int64                      `json:"created_at"`
	SendAt    int64                      `json:"send_at"` // When grace period expires
	Cancelled bool                       `json:"cancelled"`
}

// UndoSendResponse represents an undo send operation response.
type UndoSendResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	TimeLeft  int    `json:"time_left_sec,omitempty"`
}

// =============================================================================
// Email Templates Types
// =============================================================================

// EmailTemplate represents a reusable email template.
type EmailTemplate struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Subject    string            `json:"subject,omitempty"`
	Body       string            `json:"body"`
	Shortcut   string            `json:"shortcut,omitempty"`  // e.g., "/thanks", "/intro"
	Variables  []string          `json:"variables,omitempty"` // Placeholders like {{name}}, {{company}}
	Category   string            `json:"category,omitempty"`  // "greeting", "follow-up", "closing"
	UsageCount int               `json:"usage_count"`
	CreatedAt  int64             `json:"created_at"`
	UpdatedAt  int64             `json:"updated_at"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// TemplateListResponse represents a list of templates.
type TemplateListResponse struct {
	Templates []EmailTemplate `json:"templates"`
	Total     int             `json:"total"`
}
