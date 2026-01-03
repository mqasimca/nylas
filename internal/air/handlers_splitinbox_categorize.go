package air

import (
	"cmp"
	"encoding/json"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
)

// =============================================================================
// Email Categorization
// =============================================================================

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

	// Check custom rules (sorted by priority descending)
	rules := make([]CategoryRule, len(config.Rules))
	copy(rules, config.Rules)
	slices.SortFunc(rules, func(a, b CategoryRule) int {
		return cmp.Compare(b.Priority, a.Priority) // descending
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
