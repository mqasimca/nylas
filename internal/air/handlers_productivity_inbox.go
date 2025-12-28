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
	EmailID       string        `json:"email_id"`
	Category      InboxCategory `json:"category"`
	MatchedRule   string        `json:"matched_rule,omitempty"`
	CategorizedAt int64         `json:"categorized_at"`
}

// SplitInboxConfig holds the split inbox configuration.
type SplitInboxConfig struct {
	Enabled    bool            `json:"enabled"`
	Categories []InboxCategory `json:"categories"`
	VIPSenders []string        `json:"vip_senders"` // Email addresses marked as VIP
	Rules      []CategoryRule  `json:"rules"`
}

// SplitInboxResponse represents the split inbox API response.
type SplitInboxResponse struct {
	Config     SplitInboxConfig                  `json:"config"`
	Categories map[InboxCategory]int             `json:"category_counts"`
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
		EmailID string            `json:"email_id"`
		From    string            `json:"from"`
		Subject string            `json:"subject"`
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
