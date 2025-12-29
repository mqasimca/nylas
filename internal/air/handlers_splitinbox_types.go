package air

// =============================================================================
// Split Inbox Types & Constants
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

// =============================================================================
// Default Categorization Patterns
// =============================================================================

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
