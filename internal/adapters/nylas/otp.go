package nylas

import (
	"regexp"
	"strings"

	"github.com/mqasimca/nylas/internal/domain"
)

// OTP extraction patterns - ordered by specificity
var otpPatterns = []*regexp.Regexp{
	// Explicit OTP patterns
	regexp.MustCompile(`(?i)\botp[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\botp\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\botp\s+for\s+\w+\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bone[- ]?time[- ]?(?:password|code|passcode)[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bone[- ]?time[- ]?(?:password|code|passcode)\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bverification[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bverification[- ]?code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bsecurity[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bsecurity[- ]?code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bauth(?:entication)?[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\b2fa[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\b2fa[- ]?code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\btwo[- ]?factor[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bsign[- ]?in[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\blogin[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\blogin[- ]?code\s+is\s+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bpin[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bpasscode[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bconfirmation[- ]?code[:\s]+(\d{4,8})\b`),
	regexp.MustCompile(`(?i)\bconfirmation[- ]?code\s+is\s+(\d{4,8})\b`),

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
	regexp.MustCompile(`(?i)\byour\s+code\s+is\s+(\d{4,8})\b`),

	// Provider-specific patterns
	regexp.MustCompile(`(?i)\bG-(\d{6})\b`),                                               // Google
	regexp.MustCompile(`(?i)(?:microsoft|outlook)\s+.*?\bcode[:\s]+(\d{4,8})\b`),          // Microsoft
	regexp.MustCompile(`(?i)(?:github|device\s+verification)\s+.*?\bcode[:\s]+(\d{4,8})`), // GitHub
	regexp.MustCompile(`(?i)(?:amazon|aws)\s+.*?\b(\d{6})\b`),                             // Amazon
	regexp.MustCompile(`(?i)(?:google)\s+.*?\bcode\s+is\s+(\d{4,8})\b`),                   // Google

	// HTML patterns
	regexp.MustCompile(`(?i)<span[^>]*>(\d{4,8})</span>`),
	regexp.MustCompile(`(?i)<div[^>]*>(\d{4,8})</div>`),
	regexp.MustCompile(`(?i)<strong[^>]*>(\d{4,8})</strong>`),
	regexp.MustCompile(`(?i)<p[^>]*>(\d{4,8})</p>`),

	// Context-aware pattern - digits on their own line in OTP context
	regexp.MustCompile(`(?i)(?:otp|code|verification|password)\s*[:\n]\s*(\d{4,8})\b`),

	// Fallback: Find digits in body when subject indicates OTP
	regexp.MustCompile(`(?i)(?:one.?time|otp|password|verification).{0,50}?(\d{4,8})`),
}

// Year patterns to exclude
var yearPattern = regexp.MustCompile(`(?i)(?:copyright|©|\d{4}\s*-\s*\d{4}|year|date|january|february|march|april|may|june|july|august|september|october|november|december)\s*[\s:]*(\d{4})`)
var standaloneYearPattern = regexp.MustCompile(`\b(19\d{2}|20[0-2]\d)\b`)

// strongOTPSubjectPattern matches subjects that strongly indicate an OTP email
var strongOTPSubjectPattern = regexp.MustCompile(`(?i)^(?:otp|one[- ]?time[- ]?(?:pass(?:word|code)?|code)|verification[- ]?code|security[- ]?code|2fa|two[- ]?factor)$`)

// standaloneCodePattern matches standalone 3-8 digit codes
var standaloneCodePattern = regexp.MustCompile(`\b(\d{3,8})\b`)

// ExtractOTP attempts to extract an OTP code from message content.
func ExtractOTP(subject, body string) string {
	// Combine subject and body for searching
	content := subject + " " + body

	// Try each pattern
	for _, pattern := range otpPatterns {
		matches := pattern.FindStringSubmatch(content)
		if len(matches) >= 2 {
			code := matches[1]
			if isLikelyOTP(code, content) {
				return code
			}
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
		"otp", "one-time", "one time", "verification code", "verify",
		"security code", "auth code", "authentication", "2fa", "two-factor",
		"sign in", "login", "passcode", "confirm",
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
