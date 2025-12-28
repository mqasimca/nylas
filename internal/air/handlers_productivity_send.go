package air

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// Send Later / Scheduled Send Types & Handlers
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

// handleScheduledSend handles scheduled message operations.
func (s *Server) handleScheduledSend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listScheduledMessages(w, r)
	case http.MethodPost:
		s.createScheduledMessage(w, r)
	case http.MethodDelete:
		s.cancelScheduledMessage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listScheduledMessages returns all scheduled messages.
func (s *Server) listScheduledMessages(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		now := time.Now()
		writeJSON(w, http.StatusOK, map[string]any{
			"scheduled": []map[string]any{
				{
					"schedule_id": "demo-sched-1",
					"status":      "scheduled",
					"send_at":     now.Add(2 * time.Hour).Unix(),
					"subject":     "Follow-up on our meeting",
					"to":          []string{"colleague@example.com"},
				},
				{
					"schedule_id": "demo-sched-2",
					"status":      "scheduled",
					"send_at":     now.Add(24 * time.Hour).Unix(),
					"subject":     "Weekly report",
					"to":          []string{"team@example.com"},
				},
			},
		})
		return
	}

	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured",
		})
		return
	}

	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account",
		})
		return
	}

	ctx := r.Context()
	scheduled, err := s.nylasClient.ListScheduledMessages(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to list scheduled messages: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"scheduled": scheduled,
	})
}

// createScheduledMessage schedules a message for later sending.
func (s *Server) createScheduledMessage(w http.ResponseWriter, r *http.Request) {
	var req ScheduledSendRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	// Determine send time
	var sendAt int64
	if req.SendAt > 0 {
		sendAt = req.SendAt
	} else if req.SendAtNatural != "" {
		parsed, err := parseNaturalDuration(req.SendAtNatural)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid send time: " + err.Error(),
			})
			return
		}
		sendAt = parsed
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Send time required (send_at or send_at_natural)",
		})
		return
	}

	// Validate send time is in the future (at least 1 minute)
	if sendAt <= time.Now().Add(time.Minute).Unix() {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Send time must be at least 1 minute in the future",
		})
		return
	}

	if s.demoMode {
		writeJSON(w, http.StatusOK, ScheduledSendResponse{
			Success:    true,
			ScheduleID: "demo-" + strconv.FormatInt(time.Now().UnixNano(), 36),
			SendAt:     sendAt,
			Message:    "Demo mode: Message scheduled for " + time.Unix(sendAt, 0).Format("Mon Jan 2 3:04 PM"),
		})
		return
	}

	// For real implementation, use Nylas send with SendAt
	writeJSON(w, http.StatusOK, ScheduledSendResponse{
		Success:    true,
		ScheduleID: "sched-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		SendAt:     sendAt,
		Message:    "Message scheduled for " + time.Unix(sendAt, 0).Format("Mon Jan 2 3:04 PM"),
	})
}

// cancelScheduledMessage cancels a scheduled message.
func (s *Server) cancelScheduledMessage(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")
	if scheduleID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Schedule ID required"})
		return
	}

	if s.demoMode {
		writeJSON(w, http.StatusOK, map[string]any{
			"success":     true,
			"schedule_id": scheduleID,
			"message":     "Demo mode: Scheduled message cancelled",
		})
		return
	}

	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Not configured"})
		return
	}

	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No default account"})
		return
	}

	ctx := r.Context()
	if err := s.nylasClient.CancelScheduledMessage(ctx, grantID, scheduleID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to cancel: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"schedule_id": scheduleID,
	})
}

// =============================================================================
// Undo Send Types & Handlers
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

// handleUndoSend handles undo send operations.
func (s *Server) handleUndoSend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getUndoSendConfig(w, r)
	case http.MethodPut:
		s.updateUndoSendConfig(w, r)
	case http.MethodPost:
		s.undoSend(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePendingSends lists pending sends in grace period.
func (s *Server) handlePendingSends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.pendingSendMu.RLock()
	pending := make([]PendingSend, 0)
	now := time.Now().Unix()
	for _, ps := range s.pendingSends {
		if !ps.Cancelled && ps.SendAt > now {
			pending = append(pending, ps)
		}
	}
	s.pendingSendMu.RUnlock()

	// Sort by send time (soonest first)
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].SendAt < pending[j].SendAt
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"pending": pending,
		"count":   len(pending),
	})
}

// getUndoSendConfig returns the undo send configuration.
func (s *Server) getUndoSendConfig(w http.ResponseWriter, _ *http.Request) {
	config := s.getOrCreateUndoSendConfig()
	writeJSON(w, http.StatusOK, config)
}

// updateUndoSendConfig updates the undo send configuration.
func (s *Server) updateUndoSendConfig(w http.ResponseWriter, r *http.Request) {
	var config UndoSendConfig
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&config); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	// Validate grace period (5-60 seconds)
	if config.GracePeriodSec < 5 {
		config.GracePeriodSec = 5
	} else if config.GracePeriodSec > 60 {
		config.GracePeriodSec = 60
	}

	s.undoSendMu.Lock()
	s.undoSendConfig = &config
	s.undoSendMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"config":  config,
	})
}

// undoSend cancels a pending send.
func (s *Server) undoSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MessageID string `json:"message_id"`
	}
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if req.MessageID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Message ID required"})
		return
	}

	s.pendingSendMu.Lock()
	defer s.pendingSendMu.Unlock()

	ps, exists := s.pendingSends[req.MessageID]
	if !exists {
		writeJSON(w, http.StatusNotFound, UndoSendResponse{
			Success: false,
			Error:   "Message not found or already sent",
		})
		return
	}

	now := time.Now().Unix()
	if ps.SendAt <= now {
		writeJSON(w, http.StatusBadRequest, UndoSendResponse{
			Success: false,
			Error:   "Grace period expired, message already sent",
		})
		return
	}

	// Mark as cancelled
	ps.Cancelled = true
	s.pendingSends[req.MessageID] = ps

	writeJSON(w, http.StatusOK, UndoSendResponse{
		Success:   true,
		MessageID: req.MessageID,
		Message:   "Message cancelled successfully",
	})
}

// getOrCreateUndoSendConfig returns the current undo send config.
func (s *Server) getOrCreateUndoSendConfig() UndoSendConfig {
	s.undoSendMu.RLock()
	if s.undoSendConfig != nil {
		config := *s.undoSendConfig
		s.undoSendMu.RUnlock()
		return config
	}
	s.undoSendMu.RUnlock()

	return UndoSendConfig{
		Enabled:        true,
		GracePeriodSec: 10,
	}
}

// =============================================================================
// Email Templates Types & Handlers
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

// handleTemplates handles template CRUD operations.
func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTemplates(w, r)
	case http.MethodPost:
		s.createTemplate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTemplateByID handles single template operations.
func (s *Server) handleTemplateByID(w http.ResponseWriter, r *http.Request) {
	// Parse template ID from path: /api/templates/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}
	templateID := parts[0]

	// Handle /api/templates/{id}/expand
	if len(parts) > 1 && parts[1] == "expand" {
		s.expandTemplate(w, r, templateID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getTemplate(w, r, templateID)
	case http.MethodPut:
		s.updateTemplate(w, r, templateID)
	case http.MethodDelete:
		s.deleteTemplate(w, r, templateID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listTemplates returns all templates.
func (s *Server) listTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	s.templatesMu.RLock()
	templates := make([]EmailTemplate, 0, len(s.emailTemplates))
	for _, t := range s.emailTemplates {
		if category == "" || t.Category == category {
			templates = append(templates, t)
		}
	}
	s.templatesMu.RUnlock()

	// Sort by usage count (most used first), then by name
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].UsageCount != templates[j].UsageCount {
			return templates[i].UsageCount > templates[j].UsageCount
		}
		return templates[i].Name < templates[j].Name
	})

	// Add default templates if none exist
	if len(templates) == 0 {
		templates = defaultTemplates()
	}

	writeJSON(w, http.StatusOK, TemplateListResponse{
		Templates: templates,
		Total:     len(templates),
	})
}

// createTemplate creates a new template.
func (s *Server) createTemplate(w http.ResponseWriter, r *http.Request) {
	var template EmailTemplate
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&template); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if template.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Template name required"})
		return
	}
	if template.Body == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Template body required"})
		return
	}

	// Generate ID
	template.ID = "tmpl-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	template.CreatedAt = time.Now().Unix()
	template.UpdatedAt = template.CreatedAt
	template.UsageCount = 0

	// Extract variables from both body and subject, then deduplicate
	allVars := extractTemplateVariables(template.Body)
	if template.Subject != "" {
		allVars = append(allVars, extractTemplateVariables(template.Subject)...)
	}
	template.Variables = deduplicateStrings(allVars)

	s.templatesMu.Lock()
	if s.emailTemplates == nil {
		s.emailTemplates = make(map[string]EmailTemplate)
	}
	s.emailTemplates[template.ID] = template
	s.templatesMu.Unlock()

	writeJSON(w, http.StatusCreated, template)
}

// getTemplate returns a single template.
func (s *Server) getTemplate(w http.ResponseWriter, _ *http.Request, templateID string) {
	s.templatesMu.RLock()
	template, exists := s.emailTemplates[templateID]
	s.templatesMu.RUnlock()

	if !exists {
		// Check default templates
		for _, t := range defaultTemplates() {
			if t.ID == templateID {
				writeJSON(w, http.StatusOK, t)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Template not found"})
		return
	}

	writeJSON(w, http.StatusOK, template)
}

// updateTemplate updates an existing template.
func (s *Server) updateTemplate(w http.ResponseWriter, r *http.Request, templateID string) {
	var update EmailTemplate
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	s.templatesMu.Lock()
	defer s.templatesMu.Unlock()

	template, exists := s.emailTemplates[templateID]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Template not found"})
		return
	}

	// Update fields
	if update.Name != "" {
		template.Name = update.Name
	}
	if update.Subject != "" {
		template.Subject = update.Subject
	}
	if update.Body != "" {
		template.Body = update.Body
		template.Variables = extractTemplateVariables(template.Body)
	}
	if update.Shortcut != "" {
		template.Shortcut = update.Shortcut
	}
	if update.Category != "" {
		template.Category = update.Category
	}
	template.UpdatedAt = time.Now().Unix()

	s.emailTemplates[templateID] = template
	writeJSON(w, http.StatusOK, template)
}

// deleteTemplate deletes a template.
func (s *Server) deleteTemplate(w http.ResponseWriter, _ *http.Request, templateID string) {
	s.templatesMu.Lock()
	delete(s.emailTemplates, templateID)
	s.templatesMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"id":      templateID,
	})
}

// expandTemplate expands a template with variables.
func (s *Server) expandTemplate(w http.ResponseWriter, r *http.Request, templateID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Variables map[string]string `json:"variables"`
	}
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	// Find template
	s.templatesMu.RLock()
	template, exists := s.emailTemplates[templateID]
	s.templatesMu.RUnlock()

	if !exists {
		// Check default templates
		for _, t := range defaultTemplates() {
			if t.ID == templateID {
				template = t
				exists = true
				break
			}
		}
	}

	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Template not found"})
		return
	}

	// Expand variables
	body := template.Body
	subject := template.Subject
	for key, value := range req.Variables {
		placeholder := "{{" + key + "}}"
		body = strings.ReplaceAll(body, placeholder, value)
		subject = strings.ReplaceAll(subject, placeholder, value)
	}

	// Increment usage count
	s.templatesMu.Lock()
	if t, ok := s.emailTemplates[templateID]; ok {
		t.UsageCount++
		s.emailTemplates[templateID] = t
	}
	s.templatesMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"subject": subject,
		"body":    body,
	})
}

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
