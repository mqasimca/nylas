package nylas

import (
	"regexp"
	"strings"

	"github.com/mqasimca/nylas/internal/domain"
)

// OTP extraction patterns - ordered by specificity (provider-specific first, then generic)
var otpPatterns = []*regexp.Regexp{
	// ============================================
	// PROVIDER-SPECIFIC PATTERNS (highest priority)
	// ============================================

	// Google - "G-123456" format
	regexp.MustCompile(`(?i)\bG-(\d{6})\b`),
	regexp.MustCompile(`(?i)(?:google|gmail)\s+.*?\bcode[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)(?:google|gmail)\s+.*?\bcode\s+is\s+(\d{4,8})\b`),

	// Apple - "Your Apple ID Code is: 123456" and domain-bound "@apple.com #123456"
	regexp.MustCompile(`(?i)apple\s+(?:id\s+)?code[:\s]+is[:\s]+(\d{6})\b`),
	regexp.MustCompile(`(?i)@apple\.com\s+#(\d{6})\b`),
	regexp.MustCompile(`(?i)your\s+apple\s+(?:id\s+)?(?:code|verification)[:\s]+(\d{6})\b`),

	// Facebook/Meta - "FB-12345" format (5 digits)
	regexp.MustCompile(`(?i)\bFB-(\d{5,6})\b`),
	regexp.MustCompile(`(?i)(\d{5,6})\s+is\s+your\s+facebook\s+(?:confirmation\s+)?code\b`),
	regexp.MustCompile(`(?i)facebook\s+(?:confirmation\s+)?code[:\s]+(\d{5,6})\b`),

	// Instagram
	regexp.MustCompile(`(?i)(\d{6})\s+is\s+your\s+instagram\s+code\b`),
	regexp.MustCompile(`(?i)instagram\s+(?:security\s+)?code[:\s]+(\d{6})\b`),

	// WhatsApp - "123-456" or "123 456" format
	regexp.MustCompile(`(?i)whatsapp\s+code[:\s]+(\d{3}[- ]?\d{3})\b`),
	regexp.MustCompile(`(?i)whatsapp\s+code\s+is\s+(\d{3}[- ]?\d{3})\b`),
	regexp.MustCompile(`(?i)your\s+whatsapp\s+(?:.*?)?code[:\s]+(\d{3}[- ]?\d{3})\b`),
	regexp.MustCompile(`(?i)your\s+whatsapp\s+code\s+is\s+(\d{3}[- ]?\d{3})\b`),
	regexp.MustCompile(`(?i)whatsapp[:\s]+(\d{6})\b`),

	// Twitter/X
	regexp.MustCompile(`(?i)(?:twitter|x\.com)\s+(?:.*?)?code[:\s]+(\d{6,8})\b`),
	regexp.MustCompile(`(?i)(\d{6,8})\s+is\s+your\s+(?:twitter|x)\s+(?:verification\s+)?code\b`),

	// LinkedIn
	regexp.MustCompile(`(?i)linkedin\s+(?:.*?)?code[:\s]+(\d{6})\b`),
	regexp.MustCompile(`(?i)(\d{6})\s+is\s+your\s+linkedin\s+(?:verification\s+)?code\b`),

	// Microsoft/Outlook/Azure
	regexp.MustCompile(`(?i)(?:microsoft|outlook|azure|xbox|office)\s+.*?\bcode[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)(?:microsoft|outlook)\s+(?:security\s+)?code[:\s]+(\d{6,8})\b`),

	// GitHub
	regexp.MustCompile(`(?i)(?:github|device\s+verification)\s+.*?\bcode[:\s]+(\d{4,8})`),
	regexp.MustCompile(`(?i)github\s+(?:verification\s+)?code[:\s]+(\d{6,8})\b`),

	// Amazon/AWS (requires OTP context to avoid false positives)
	regexp.MustCompile(`(?i)(?:amazon|aws)\s+(?:otp|verification|security|sign[- ]?in)\s+.*?\b(\d{6})\b`),
	regexp.MustCompile(`(?i)amazon\s+(?:.*?)?code[:\s]+(\d{6})\b`),

	// Slack
	regexp.MustCompile(`(?i)slack\s+(?:.*?)?code[:\s]+(\d{6})\b`),
	regexp.MustCompile(`(?i)(\d{6})\s+is\s+your\s+slack\s+(?:verification\s+)?code\b`),

	// Discord
	regexp.MustCompile(`(?i)discord\s+(?:.*?)?code[:\s]+(\d{6})\b`),
	regexp.MustCompile(`(?i)(\d{6})\s+is\s+your\s+discord\s+(?:verification\s+)?code\b`),

	// Telegram
	regexp.MustCompile(`(?i)telegram\s+(?:.*?)?code[:\s]+(\d{5,6})\b`),
	regexp.MustCompile(`(?i)(\d{5,6})\s+is\s+your\s+telegram\s+(?:login\s+)?code\b`),

	// Signal
	regexp.MustCompile(`(?i)signal\s+(?:.*?)?code[:\s]+(\d{6})\b`),

	// Uber/Lyft/DoorDash/Grubhub (ride-sharing & delivery)
	regexp.MustCompile(`(?i)(?:uber|lyft|doordash|grubhub|postmates|instacart)\s+code[:\s]+(\d{4,6})\b`),
	regexp.MustCompile(`(?i)your\s+(?:uber|lyft|doordash|grubhub)\s+code\s+is[:\s]+(\d{4,6})\b`),
	regexp.MustCompile(`(?i)(\d{4,6})\s+is\s+your\s+(?:uber|lyft|doordash|grubhub)\s+code\b`),

	// Airbnb/Booking.com/VRBO (travel)
	regexp.MustCompile(`(?i)(?:airbnb|booking\.com|vrbo|expedia)\s+(?:.*?)?code[:\s]+(\d{4,6})\b`),
	regexp.MustCompile(`(?i)(\d{4,6})\s+is\s+your\s+(?:airbnb|booking)\s+(?:verification\s+)?code\b`),

	// Netflix/Spotify/Disney+ (streaming)
	regexp.MustCompile(`(?i)(?:netflix|spotify|disney\+?|hulu|hbo|prime\s+video)\s+(?:.*?)?code[:\s]+(\d{4,6})\b`),

	// PayPal/Venmo/CashApp/Zelle (payments)
	regexp.MustCompile(`(?i)(?:paypal|venmo|cash\s*app|zelle|wise|revolut)\s+(?:.*?)?code[:\s]+(\d{4,6})\b`),
	regexp.MustCompile(`(?i)(\d{4,6})\s+is\s+your\s+(?:paypal|venmo|cash\s*app)\s+(?:security\s+)?code\b`),

	// Coinbase/Binance/Kraken (crypto)
	regexp.MustCompile(`(?i)(?:coinbase|binance|kraken|gemini|crypto\.com)\s+(?:.*?)?code[:\s]+(\d{6,8})\b`),
	regexp.MustCompile(`(?i)(\d{6,8})\s+is\s+your\s+(?:coinbase|binance|kraken)\s+(?:verification\s+)?code\b`),

	// Zoom/Teams/Webex (video conferencing)
	regexp.MustCompile(`(?i)(?:zoom|teams|webex|meet)\s+(?:.*?)?code[:\s]+(\d{6})\b`),

	// Shopify/Etsy/eBay (e-commerce)
	regexp.MustCompile(`(?i)(?:shopify|etsy|ebay|amazon)\s+(?:seller\s+)?(?:.*?)?code[:\s]+(\d{6})\b`),

	// Dropbox/Box/OneDrive (cloud storage)
	regexp.MustCompile(`(?i)(?:dropbox|box|onedrive|icloud)\s+(?:.*?)?code[:\s]+(\d{6})\b`),

	// Atlassian/Jira/Confluence
	regexp.MustCompile(`(?i)(?:atlassian|jira|confluence|bitbucket)\s+(?:.*?)?code[:\s]+(\d{6})\b`),

	// Okta/Auth0/OneLogin (SSO providers)
	regexp.MustCompile(`(?i)(?:okta|auth0|onelogin|duo)\s+(?:.*?)?code[:\s]+(\d{6})\b`),

	// ============================================
	// GENERIC OTP PATTERNS
	// ============================================

	// Explicit OTP patterns
	regexp.MustCompile(`(?i)\botp[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\botp\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\botp\s+for\s+\w+\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bone[- ]?time[- ]?(?:password|code|passcode)[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bone[- ]?time[- ]?(?:password|code|passcode)\s+is\s+(\d{4,8})\b`),

	// Verification code patterns
	regexp.MustCompile(`(?i)\bverification[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bverification[- ]?code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bverify[- ]?code[:\s]+(\d{4,8})\b`),

	// Security code patterns
	regexp.MustCompile(`(?i)\bsecurity[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bsecurity[- ]?code\s+is\s+(\d{4,8})\b`),

	// Authentication code patterns
	regexp.MustCompile(`(?i)\bauth(?:entication)?[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\baccess[- ]?code[:\s]+(\d{4,8})\b`),

	// 2FA/MFA patterns
	regexp.MustCompile(`(?i)\b2fa[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\b2fa[- ]?code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bmfa[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\btwo[- ]?factor[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bmulti[- ]?factor[- ]?code[:\s]+(\d{4,8})\b`),

	// Sign-in/Login patterns
	regexp.MustCompile(`(?i)\bsign[- ]?in[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\blogin[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\blogin[- ]?code\s+is\s+(\d{4,8})\b`),

	// PIN/Passcode patterns
	regexp.MustCompile(`(?i)\bpin[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bpasscode[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\btemporary[- ]?(?:pin|password|code)[:\s]+(\d{4,8})\b`),

	// Confirmation code patterns
	regexp.MustCompile(`(?i)\bconfirmation[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bconfirmation[- ]?code\s+is\s+(\d{4,8})\b`),

	// Reset/Recovery patterns
	regexp.MustCompile(`(?i)\breset[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\brecovery[- ]?code[:\s]+(\d{4,8})\b`),

	// Activation patterns
	regexp.MustCompile(`(?i)\bactivation[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bactivate[:\s]+(\d{4,8})\b`),

	// "code is X" patterns
	regexp.MustCompile(`(?i)\bcode[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bcode\s+is[:\s]+(\d{4,8})\b`),

	// "enter/use" patterns
	regexp.MustCompile(`(?i)\benter[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\buse[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bcode[:\s]+to\s+complete[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bconfirm\s+with[:\s]+(\d{4,8})\b`),

	// "X is your code" patterns
	regexp.MustCompile(`(?i)\b(\d{4,8})\s+is\s+your\s+(?:verification\s+)?code\b`),
	regexp.MustCompile(`(?i)\b(\d{4,8})\s+is\s+your\s+otp\b`),
	regexp.MustCompile(`(?i)\b(\d{4,8})\s+is\s+your\s+one[- ]?time[- ]?(?:password|code)\b`),
	regexp.MustCompile(`(?i)\b(\d{4,8})\s+is\s+your\s+(?:security|confirmation|verification)\s+code\b`),
	regexp.MustCompile(`(?i)\byour\s+code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\byour\s+(?:verification|security|confirmation)\s+code\s+is\s+(\d{4,8})\b`),

	// ============================================
	// HTML PATTERNS (for email body parsing)
	// ============================================
	regexp.MustCompile(`(?i)<span[^>]*>(\d{4,8})</span>`),
	regexp.MustCompile(`(?i)<div[^>]*>(\d{4,8})</div>`),
	regexp.MustCompile(`(?i)<strong[^>]*>(\d{4,8})</strong>`),
	regexp.MustCompile(`(?i)<b[^>]*>(\d{4,8})</b>`),
	regexp.MustCompile(`(?i)<p[^>]*>(\d{4,8})</p>`),
	regexp.MustCompile(`(?i)<td[^>]*>(\d{4,8})</td>`),
	regexp.MustCompile(`(?i)<h\d[^>]*>(\d{4,8})</h\d>`),
	regexp.MustCompile(`(?i)class="[^"]*(?:otp|code|verification)[^"]*"[^>]*>(\d{4,8})<`),

	// Spaced digit patterns are handled separately in ExtractOTP function

	// Context-aware pattern - digits on their own line in OTP context
	regexp.MustCompile(`(?i)(?:otp|code|verification|password)\s*[:\n]\s*(\d{4,8})\b`),

	// Fallback: Find digits in body when subject indicates OTP
	regexp.MustCompile(`(?i)(?:one.?time|otp|password|verification).{0,50}?(\d{4,8})`),
}

// standaloneYearPattern excludes years from OTP matching
var standaloneYearPattern = regexp.MustCompile(`\b(19\d{2}|20[0-2]\d)\b`)

// strongOTPSubjectPattern matches subjects that strongly indicate an OTP email
var strongOTPSubjectPattern = regexp.MustCompile(`(?i)^(?:otp|one[- ]?time[- ]?(?:pass(?:word|code)?|code)|verification[- ]?code|security[- ]?code|2fa|two[- ]?factor)$`)

// standaloneCodePattern matches standalone 3-8 digit codes
var standaloneCodePattern = regexp.MustCompile(`\b(\d{3,8})\b`)

// spacedDigitPattern matches spaced digits like "1 2 3 4 5 6" in OTP context
var spacedDigitPattern = regexp.MustCompile(`(?i)(?:code|otp|verification)[:\s]+(\d\s+\d\s+\d\s+\d(?:\s+\d)?(?:\s+\d)?)`)

// ExtractOTP attempts to extract an OTP code from message content.
func ExtractOTP(subject, body string) string {
	// Combine subject and body for searching
	content := subject + " " + body

	// Try each pattern
	for _, pattern := range otpPatterns {
		matches := pattern.FindStringSubmatch(content)
		if len(matches) >= 2 {
			code := matches[1]
			// Handle WhatsApp-style codes with hyphens (123-456 -> 123456)
			code = strings.ReplaceAll(code, "-", "")
			code = strings.ReplaceAll(code, " ", "")
			if isLikelyOTP(code, content) {
				return code
			}
		}
	}

	// Try spaced digit pattern (e.g., "1 2 3 4 5 6" -> "123456")
	if matches := spacedDigitPattern.FindStringSubmatch(content); len(matches) >= 2 {
		code := strings.ReplaceAll(matches[1], " ", "")
		if len(code) >= 4 && len(code) <= 8 && isLikelyOTP(code, content) {
			return code
		}
	}

	// If subject strongly indicates OTP, look for any standalone 3-8 digit code in body
	if strongOTPSubjectPattern.MatchString(strings.TrimSpace(subject)) {
		matches := standaloneCodePattern.FindStringSubmatch(body)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// isLikelyOTP checks if the code is likely an OTP and not a year or other number.
func isLikelyOTP(code, content string) bool {
	// Exclude years
	if len(code) == 4 {
		year := code
		if year >= "1900" && year <= "2099" {
			// Check if it's used in a year context
			lowerContent := strings.ToLower(content)
			if strings.Contains(lowerContent, "copyright") ||
				strings.Contains(lowerContent, "©") ||
				strings.Contains(lowerContent, year+"-") ||
				strings.Contains(lowerContent, "-"+year) {
				return false
			}

			// If there's clear OTP context, allow it
			if hasOTPContext(lowerContent) {
				return true
			}

			// Otherwise, be conservative about years
			if standaloneYearPattern.MatchString(code) {
				return false
			}
		}
	}

	return true
}

// hasOTPContext checks if the content has OTP-related keywords.
func hasOTPContext(content string) bool {
	otpKeywords := []string{
		// Core OTP terms
		"otp", "one-time", "one time", "verification code", "verify",
		"security code", "auth code", "authentication", "2fa", "two-factor",
		"mfa", "multi-factor", "sign in", "login", "passcode", "confirm",
		// Action terms
		"enter code", "enter the code", "use code", "use this code",
		"confirmation code", "access code", "temporary code", "temporary password",
		"reset code", "recovery code", "activation code",
		// Provider hints
		"don't share", "do not share", "expires in", "valid for",
		"this code will expire", "never share this",
	}

	for _, keyword := range otpKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// FindOTP searches messages for an OTP and returns the result.
func FindOTP(messages []domain.Message) (*domain.OTPResult, error) {
	for _, msg := range messages {
		code := ExtractOTP(msg.Subject, msg.Body)
		if code == "" {
			// Try snippet if body didn't work
			code = ExtractOTP(msg.Subject, msg.Snippet)
		}
		if code != "" {
			from := ""
			if len(msg.From) > 0 {
				from = msg.From[0].Email
			}
			return &domain.OTPResult{
				Code:      code,
				From:      from,
				Subject:   msg.Subject,
				Received:  msg.Date,
				MessageID: msg.ID,
			}, nil
		}
	}
	return nil, domain.ErrOTPNotFound
}
