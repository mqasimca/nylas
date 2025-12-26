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
// Split Inbox Types & Handlers
// =============================================================================

// InboxCategory represents an email category for split inbox.
type InboxCategory string

const (
	CategoryPrimary     InboxCategory = "primary"
	CategoryVIP         InboxCategory = "vip"
	CategoryNewsletters InboxCategory = "newsletters"
	CategoryUpdates     InboxCategory = "updates"
	CategorySocial      InboxCategory = "social"
	CategoryPromotions  InboxCategory = "promotions"
	CategoryForums      InboxCategory = "forums"
)

// CategoryRule defines a rule for categorizing emails.
type CategoryRule struct {
	ID          string        `json:"id"`
	Category    InboxCategory `json:"category"`
	Type        string        `json:"type"` // "sender", "domain", "subject", "header"
	Pattern     string        `json:"pattern"`
	IsRegex     bool          `json:"is_regex"`
	Priority    int           `json:"priority"` // Higher priority rules are checked first
	Description string        `json:"description,omitempty"`
	CreatedAt   int64         `json:"created_at"`
}

// CategorizedEmail represents an email with its category.
type CategorizedEmail struct {
	EmailID      string        `json:"email_id"`
	Category     InboxCategory `json:"category"`
	MatchedRule  string        `json:"matched_rule,omitempty"`
	CategorizedAt int64        `json:"categorized_at"`
}

// SplitInboxConfig holds the split inbox configuration.
type SplitInboxConfig struct {
	Enabled    bool           `json:"enabled"`
	Categories []InboxCategory `json:"categories"`
	VIPSenders []string       `json:"vip_senders"` // Email addresses marked as VIP
	Rules      []CategoryRule `json:"rules"`
}

// SplitInboxResponse represents the split inbox API response.
type SplitInboxResponse struct {
	Config     SplitInboxConfig            `json:"config"`
	Categories map[InboxCategory]int       `json:"category_counts"`
	Recent     map[InboxCategory][]EmailResponse `json:"recent,omitempty"`
}

// Default newsletter patterns.
var defaultNewsletterPatterns = []string{
	"noreply@", "newsletter@", "updates@", "digest@", "news@",
	"notifications@", "mailer@", "info@", "no-reply@",
	"unsubscribe", "list-unsubscribe",
}

// Default social patterns.
var defaultSocialPatterns = []string{
	"@facebook.com", "@twitter.com", "@x.com", "@linkedin.com",
	"@instagram.com", "@tiktok.com", "@pinterest.com",
	"facebookmail.com", "linkedin.com",
}

// Default promotion patterns.
var defaultPromotionPatterns = []string{
	"deals@", "offers@", "promo@", "sale@", "discount@",
	"marketing@", "promotions@", "special@",
}

// Default update patterns (transactional).
var defaultUpdatePatterns = []string{
	"order@", "receipt@", "shipping@", "delivery@", "tracking@",
	"confirmation@", "booking@", "reservation@", "invoice@",
	"payment@", "billing@", "account@", "security@", "alert@",
}

// handleSplitInbox handles split inbox configuration and retrieval.
func (s *Server) handleSplitInbox(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSplitInboxConfig(w, r)
	case http.MethodPut:
		s.updateSplitInboxConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getSplitInboxConfig returns the split inbox configuration.
func (s *Server) getSplitInboxConfig(w http.ResponseWriter, _ *http.Request) {
	config := s.getOrCreateSplitInboxConfig()

	// Count emails per category (demo mode or from cache)
	counts := make(map[InboxCategory]int)
	counts[CategoryPrimary] = 50
	counts[CategoryVIP] = 5
	counts[CategoryNewsletters] = 20
	counts[CategoryUpdates] = 15
	counts[CategorySocial] = 8
	counts[CategoryPromotions] = 12

	writeJSON(w, http.StatusOK, SplitInboxResponse{
		Config:     config,
		Categories: counts,
	})
}

// updateSplitInboxConfig updates the split inbox configuration.
func (s *Server) updateSplitInboxConfig(w http.ResponseWriter, r *http.Request) {
	var config SplitInboxConfig
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&config); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	s.splitInboxMu.Lock()
	s.splitInboxConfig = &config
	s.splitInboxMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"config":  config,
	})
}

// handleCategorizeEmail categorizes a single email.
func (s *Server) handleCategorizeEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EmailID string `json:"email_id"`
		From    string `json:"from"`
		Subject string `json:"subject"`
		Headers map[string]string `json:"headers,omitempty"`
	}
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	category, rule := s.categorizeEmail(req.From, req.Subject, req.Headers)

	writeJSON(w, http.StatusOK, CategorizedEmail{
		EmailID:       req.EmailID,
		Category:      category,
		MatchedRule:   rule,
		CategorizedAt: time.Now().Unix(),
	})
}

// handleVIPSenders manages VIP sender list.
func (s *Server) handleVIPSenders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := s.getOrCreateSplitInboxConfig()
		writeJSON(w, http.StatusOK, map[string]any{
			"vip_senders": config.VIPSenders,
		})
	case http.MethodPost:
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
			return
		}
		s.addVIPSender(req.Email)
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "email": req.Email})
	case http.MethodDelete:
		email := r.URL.Query().Get("email")
		if email == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email required"})
			return
		}
		s.removeVIPSender(email)
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "email": email})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// categorizeEmail determines the category for an email.
func (s *Server) categorizeEmail(from, subject string, headers map[string]string) (InboxCategory, string) {
	config := s.getOrCreateSplitInboxConfig()
	fromLower := strings.ToLower(from)
	subjectLower := strings.ToLower(subject)

	// Check VIP first (highest priority)
	for _, vip := range config.VIPSenders {
		if strings.Contains(fromLower, strings.ToLower(vip)) {
			return CategoryVIP, "vip:" + vip
		}
	}

	// Check custom rules (sorted by priority)
	rules := make([]CategoryRule, len(config.Rules))
	copy(rules, config.Rules)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	for _, rule := range rules {
		if s.matchesRule(rule, fromLower, subjectLower, headers) {
			return rule.Category, rule.ID
		}
	}

	// Check default patterns
	if category, rule := s.matchDefaultPatterns(fromLower, subjectLower, headers); category != "" {
		return category, rule
	}

	return CategoryPrimary, "default"
}

// matchesRule checks if an email matches a category rule.
func (s *Server) matchesRule(rule CategoryRule, from, subject string, headers map[string]string) bool {
	var target string
	switch rule.Type {
	case "sender":
		target = from
	case "subject":
		target = subject
	case "domain":
		// Extract domain from email
		if idx := strings.Index(from, "@"); idx >= 0 {
			target = from[idx:]
		}
	case "header":
		// Check specific header
		for k, v := range headers {
			if strings.EqualFold(k, rule.Pattern) {
				target = v
				break
			}
		}
	default:
		target = from + " " + subject
	}

	if rule.IsRegex {
		if re, err := regexp.Compile(rule.Pattern); err == nil {
			return re.MatchString(target)
		}
		return false
	}
	return strings.Contains(target, strings.ToLower(rule.Pattern))
}

// matchDefaultPatterns checks against default category patterns.
func (s *Server) matchDefaultPatterns(from, subject string, headers map[string]string) (InboxCategory, string) {
	// Check social FIRST (specific domains take priority over generic patterns)
	for _, pattern := range defaultSocialPatterns {
		if strings.Contains(from, pattern) {
			return CategorySocial, "pattern:social:" + pattern
		}
	}

	// Check for list-unsubscribe header (strong newsletter signal)
	if _, ok := headers["List-Unsubscribe"]; ok {
		return CategoryNewsletters, "header:list-unsubscribe"
	}

	// Check newsletters
	for _, pattern := range defaultNewsletterPatterns {
		if strings.Contains(from, pattern) || strings.Contains(subject, pattern) {
			return CategoryNewsletters, "pattern:newsletter:" + pattern
		}
	}

	// Check promotions
	for _, pattern := range defaultPromotionPatterns {
		if strings.Contains(from, pattern) {
			return CategoryPromotions, "pattern:promo:" + pattern
		}
	}

	// Check updates (transactional)
	for _, pattern := range defaultUpdatePatterns {
		if strings.Contains(from, pattern) {
			return CategoryUpdates, "pattern:update:" + pattern
		}
	}

	return "", ""
}

// getOrCreateSplitInboxConfig returns the current split inbox config.
func (s *Server) getOrCreateSplitInboxConfig() SplitInboxConfig {
	s.splitInboxMu.RLock()
	if s.splitInboxConfig != nil {
		config := *s.splitInboxConfig
		s.splitInboxMu.RUnlock()
		return config
	}
	s.splitInboxMu.RUnlock()

	// Create default config
	return SplitInboxConfig{
		Enabled: true,
		Categories: []InboxCategory{
			CategoryPrimary, CategoryVIP, CategoryNewsletters,
			CategoryUpdates, CategorySocial, CategoryPromotions,
		},
		VIPSenders: []string{},
		Rules:      []CategoryRule{},
	}
}

// addVIPSender adds an email to the VIP list.
func (s *Server) addVIPSender(email string) {
	s.splitInboxMu.Lock()
	defer s.splitInboxMu.Unlock()

	if s.splitInboxConfig == nil {
		// Create default config inline to avoid deadlock
		s.splitInboxConfig = &SplitInboxConfig{
			Enabled: true,
			Categories: []InboxCategory{
				CategoryPrimary, CategoryVIP, CategoryNewsletters,
				CategoryUpdates, CategorySocial, CategoryPromotions,
			},
			VIPSenders: []string{},
			Rules:      []CategoryRule{},
		}
	}

	// Check if already exists
	for _, vip := range s.splitInboxConfig.VIPSenders {
		if strings.EqualFold(vip, email) {
			return
		}
	}
	s.splitInboxConfig.VIPSenders = append(s.splitInboxConfig.VIPSenders, email)
}

// removeVIPSender removes an email from the VIP list.
func (s *Server) removeVIPSender(email string) {
	s.splitInboxMu.Lock()
	defer s.splitInboxMu.Unlock()

	if s.splitInboxConfig == nil {
		return
	}

	filtered := make([]string, 0, len(s.splitInboxConfig.VIPSenders))
	for _, vip := range s.splitInboxConfig.VIPSenders {
		if !strings.EqualFold(vip, email) {
			filtered = append(filtered, vip)
		}
	}
	s.splitInboxConfig.VIPSenders = filtered
}

// =============================================================================
// Snooze Types & Handlers
// =============================================================================

// SnoozedEmail represents a snoozed email.
type SnoozedEmail struct {
	EmailID     string `json:"email_id"`
	SnoozeUntil int64  `json:"snooze_until"` // Unix timestamp
	OriginalFolder string `json:"original_folder,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// SnoozeRequest represents a request to snooze an email.
type SnoozeRequest struct {
	EmailID     string `json:"email_id"`
	SnoozeUntil int64  `json:"snooze_until,omitempty"`    // Explicit Unix timestamp
	Duration    string `json:"duration,omitempty"`        // Natural language: "1h", "2d", "tomorrow 9am"
}

// SnoozeResponse represents a snooze operation response.
type SnoozeResponse struct {
	Success     bool   `json:"success"`
	EmailID     string `json:"email_id"`
	SnoozeUntil int64  `json:"snooze_until"`
	Message     string `json:"message,omitempty"`
}

// handleSnooze handles snooze operations.
func (s *Server) handleSnooze(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSnoozedEmails(w, r)
	case http.MethodPost:
		s.snoozeEmail(w, r)
	case http.MethodDelete:
		s.unsnoozeEmail(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listSnoozedEmails returns all snoozed emails.
func (s *Server) listSnoozedEmails(w http.ResponseWriter, _ *http.Request) {
	s.snoozeMu.RLock()
	snoozed := make([]SnoozedEmail, 0, len(s.snoozedEmails))
	now := time.Now().Unix()
	for _, se := range s.snoozedEmails {
		if se.SnoozeUntil > now {
			snoozed = append(snoozed, se)
		}
	}
	s.snoozeMu.RUnlock()

	// Sort by snooze time (soonest first)
	sort.Slice(snoozed, func(i, j int) bool {
		return snoozed[i].SnoozeUntil < snoozed[j].SnoozeUntil
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"snoozed": snoozed,
		"count":   len(snoozed),
	})
}

// snoozeEmail snoozes an email until a specific time.
func (s *Server) snoozeEmail(w http.ResponseWriter, r *http.Request) {
	var req SnoozeRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if req.EmailID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email ID required"})
		return
	}

	var snoozeUntil int64
	if req.SnoozeUntil > 0 {
		snoozeUntil = req.SnoozeUntil
	} else if req.Duration != "" {
		parsed, err := parseNaturalDuration(req.Duration)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid duration: " + err.Error(),
			})
			return
		}
		snoozeUntil = parsed
	} else {
		// Default: snooze for 1 hour
		snoozeUntil = time.Now().Add(time.Hour).Unix()
	}

	// Validate snooze time is in the future
	if snoozeUntil <= time.Now().Unix() {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Snooze time must be in the future",
		})
		return
	}

	snoozed := SnoozedEmail{
		EmailID:     req.EmailID,
		SnoozeUntil: snoozeUntil,
		CreatedAt:   time.Now().Unix(),
	}

	s.snoozeMu.Lock()
	if s.snoozedEmails == nil {
		s.snoozedEmails = make(map[string]SnoozedEmail)
	}
	s.snoozedEmails[req.EmailID] = snoozed
	s.snoozeMu.Unlock()

	writeJSON(w, http.StatusOK, SnoozeResponse{
		Success:     true,
		EmailID:     req.EmailID,
		SnoozeUntil: snoozeUntil,
		Message:     "Email snoozed until " + time.Unix(snoozeUntil, 0).Format("Mon Jan 2 3:04 PM"),
	})
}

// unsnoozeEmail removes the snooze from an email.
func (s *Server) unsnoozeEmail(w http.ResponseWriter, r *http.Request) {
	emailID := r.URL.Query().Get("email_id")
	if emailID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email ID required"})
		return
	}

	s.snoozeMu.Lock()
	delete(s.snoozedEmails, emailID)
	s.snoozeMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"email_id": emailID,
	})
}

// parseNaturalDuration parses natural language duration into Unix timestamp.
func parseNaturalDuration(input string) (int64, error) {
	now := time.Now()
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle relative durations: "1h", "2d", "30m", "1w"
	if matched, _ := regexp.MatchString(`^\d+[hdwm]$`, input); matched {
		num, _ := strconv.Atoi(input[:len(input)-1])
		unit := input[len(input)-1]
		switch unit {
		case 'h':
			return now.Add(time.Duration(num) * time.Hour).Unix(), nil
		case 'd':
			return now.Add(time.Duration(num) * 24 * time.Hour).Unix(), nil
		case 'w':
			return now.Add(time.Duration(num) * 7 * 24 * time.Hour).Unix(), nil
		case 'm':
			return now.Add(time.Duration(num) * time.Minute).Unix(), nil
		}
	}

	// Handle "later today" (4 hours or 5 PM, whichever is first)
	if input == "later" || input == "later today" {
		later := now.Add(4 * time.Hour)
		fivePM := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location())
		if fivePM.After(now) && fivePM.Before(later) {
			return fivePM.Unix(), nil
		}
		return later.Unix(), nil
	}

	// Handle "tonight" (8 PM today)
	if input == "tonight" {
		tonight := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
		if tonight.Before(now) {
			tonight = tonight.Add(24 * time.Hour)
		}
		return tonight.Unix(), nil
	}

	// Handle "tomorrow" (9 AM tomorrow)
	if strings.HasPrefix(input, "tomorrow") {
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())

		// Check for time specification: "tomorrow 2pm", "tomorrow at 3:30"
		parts := strings.Fields(input)
		if len(parts) > 1 {
			timeStr := parts[len(parts)-1]
			if strings.HasPrefix(parts[1], "at") && len(parts) > 2 {
				timeStr = parts[2]
			}
			if hour, min, ok := parseTimeString(timeStr); ok {
				tomorrow = time.Date(now.Year(), now.Month(), now.Day()+1, hour, min, 0, 0, now.Location())
			}
		}
		return tomorrow.Unix(), nil
	}

	// Handle "next week" (Monday 9 AM)
	if input == "next week" || input == "monday" {
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextMonday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 9, 0, 0, 0, now.Location())
		return nextMonday.Unix(), nil
	}

	// Handle "this weekend" (Saturday 10 AM)
	if input == "weekend" || input == "this weekend" || input == "saturday" {
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7
		}
		saturday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSaturday, 10, 0, 0, 0, now.Location())
		return saturday.Unix(), nil
	}

	// Handle specific times: "9am", "14:30", "3:30pm"
	if hour, min, ok := parseTimeString(input); ok {
		target := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}
		return target.Unix(), nil
	}

	return 0, &parseError{input: input}
}

// parseTimeString parses time strings like "9am", "14:30", "3:30pm".
func parseTimeString(s string) (hour, min int, ok bool) {
	s = strings.ToLower(strings.TrimSpace(s))

	isPM := strings.HasSuffix(s, "pm")
	isAM := strings.HasSuffix(s, "am")
	s = strings.TrimSuffix(strings.TrimSuffix(s, "pm"), "am")

	parts := strings.Split(s, ":")
	if len(parts) == 1 {
		// Just hour: "9", "14"
		h, err := strconv.Atoi(parts[0])
		if err != nil || h < 0 || h > 23 {
			return 0, 0, false
		}
		hour = h
		min = 0
	} else if len(parts) == 2 {
		// Hour:min: "9:30", "14:00"
		h, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
			return 0, 0, false
		}
		hour = h
		min = m
	} else {
		return 0, 0, false
	}

	// Handle AM/PM
	if isPM && hour < 12 {
		hour += 12
	} else if isAM && hour == 12 {
		hour = 0
	}

	return hour, min, true
}

type parseError struct {
	input string
}

func (e *parseError) Error() string {
	return "cannot parse duration: " + e.input
}

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
	Enabled       bool `json:"enabled"`
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
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Subject     string            `json:"subject,omitempty"`
	Body        string            `json:"body"`
	Shortcut    string            `json:"shortcut,omitempty"` // e.g., "/thanks", "/intro"
	Variables   []string          `json:"variables,omitempty"` // Placeholders like {{name}}, {{company}}
	Category    string            `json:"category,omitempty"`  // "greeting", "follow-up", "closing"
	UsageCount  int               `json:"usage_count"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
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
