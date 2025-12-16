package nylas_test

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestExtractOTP(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		body    string
		want    string
	}{
		{"OTP with explicit label", "", "Your OTP is 123456", "123456"},
		{"OTP with colon", "", "OTP: 654321", "654321"},
		{"One-time password explicit", "", "One-time password: 987654", "987654"},
		{"One-time code with hyphen", "", "one-time code 112233", "112233"},
		{"Verification code with colon", "", "Verification code: 445566", "445566"},
		{"Verification code is", "", "Your verification code is 778899", "778899"},
		{"Security code colon", "", "Security code: 111222", "111222"},
		{"Security code is", "", "Your security code is 333444", "333444"},
		{"Auth code pattern", "", "Auth code: 555666", "555666"},
		{"Authentication code", "", "Authentication code: 777888", "777888"},
		{"Google G-prefix code", "", "G-123456 is your verification code", "123456"},
		{"Google code in subject", "G-789012", "", "789012"},
		{"Google 2FA code", "", "Your Google 2FA code is 345678", "345678"},
		{"Microsoft security code", "", "Microsoft security code: 901234", "901234"},
		{"Microsoft verification", "", "Microsoft verification code is 567890", "567890"},
		{"GitHub device verification", "", "GitHub device verification code: 123789", "123789"},
		{"GitHub style verification", "", "Your GitHub verification code is 456012", "456012"},
		{"Amazon OTP", "", "Amazon OTP: 789345", "789345"},
		{"Amazon verification", "", "Your Amazon verification code is 012678", "012678"},
		{"2FA code", "", "Your 2FA code is 345901", "345901"},
		{"Two-factor code", "", "Two-factor code: 678234", "678234"},
		{"Sign in code", "", "Sign in code: 901567", "901567"},
		{"Login code", "", "Your login code is 234890", "234890"},
		{"PIN code 4 digits", "", "PIN code: 1234", "1234"},
		{"PIN code 6 digits", "", "PIN: 123456", "123456"},
		{"Passcode pattern", "", "Your passcode: 654321", "654321"},
		{"Enter code pattern", "", "Enter: 987654", "987654"},
		{"Use code pattern", "", "Use: 321654", "321654"},
		{"Code to complete", "", "Code to complete: 654987", "654987"},
		{"Confirm with code", "", "Confirm with: 789456", "789456"},
		{"X is your code", "", "123456 is your code", "123456"},
		{"X is your OTP", "", "654321 is your OTP", "654321"},
		{"X is your one-time code", "", "987654 is your one-time code", "987654"},
		{"Code in HTML span", "", "<span>123456</span>", "123456"},
		{"Code in HTML div with OTP context", "", "Your OTP is <div class=\"code\">654321</div>", "654321"},
		{"Code in HTML strong with OTP context", "", "Verification code: <strong>987654</strong>", "987654"},
		{"Full email body - short", "Your OTP", "Your verification code is 123456. Do not share.", "123456"},
		{"Email with multiple numbers context", "", "Order #12345. OTP: 654321. Thank you.", "654321"},
		{"Code on its own line in OTP context", "", "Your OTP code:\n123456\nThank you", "123456"},
		{"No OTP - regular email", "", "Hello, how are you today?", ""},
		{"No OTP - order number", "", "Your order #1234567890 has shipped", ""},
		{"Year should not match - 2024", "", "Copyright 2024 Company Inc.", ""},
		{"Year should not match - 2023", "", "Year 2023 was great", ""},
		{"Year should not match - 1999", "", "Since 1999 we have been...", ""},
		{"Generic business email no OTP", "", "Meeting at 3pm. Call 555-1234 for details.", ""},
		{"Should not match phone in signature without OTP context", "", "Best regards, Call us: 1234567890", ""},
		{"4-digit OTP explicit", "", "Your OTP is 1234", "1234"},
		{"4-digit PIN", "", "PIN: 5678", "5678"},
		{"8-digit security code", "", "Security code: 12345678", "12345678"},
		{"8-digit OTP", "", "Your OTP: 87654321", "87654321"},
		{"Lowercase otp", "", "otp: 123456", "123456"},
		{"Uppercase OTP", "", "OTP: 654321", "654321"},
		{"Mixed case verification", "", "VeRiFiCaTiOn code: 111222", "111222"},
		{"Multiple spaces", "", "OTP   :   333444", "333444"},
		{"Tab separated", "", "OTP:\t555666", "555666"},
		{"Subject: One time password (common format)", "Your one time password", "Code: 123456", "123456"},
		{"Subject line with code inline", "Your code is 654321", "", "654321"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nylas.ExtractOTP(tt.subject, tt.body)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractOTPRealWorldEmails(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		body    string
		want    string
	}{
		{
			"Google verification email",
			"Your Google verification code",
			"G-123456 is your Google verification code. Don't share this code with anyone.",
			"123456",
		},
		{
			"Microsoft security code email",
			"Microsoft account security code",
			"Use 654321 as your Microsoft account security code. If you didn't request this code, you can ignore this email.",
			"654321",
		},
		{
			"GitHub device verification",
			"[GitHub] Your verification code",
			"Your GitHub verification code is 789012. Enter this code to complete your sign in.",
			"789012",
		},
		{
			"Amazon OTP email",
			"Amazon: Your verification code",
			"Your Amazon verification code is 345678. Don't share this code with anyone.",
			"345678",
		},
		{
			"Generic 2FA email",
			"Your login verification code",
			"Your verification code is 901234. This code expires in 10 minutes.",
			"901234",
		},
		{
			"Slack verification",
			"Slack confirmation code",
			"Your Slack confirmation code is 567890. Enter it in the app to verify your identity.",
			"567890",
		},
		{
			"Discord verification",
			"Your verification code",
			"Your Discord verification code is 123789. This code will expire in 10 minutes.",
			"123789",
		},
		{
			"Banking OTP with sensitive info warning",
			"One-time password for your transaction",
			"Your OTP for transaction is 456012. Never share this with anyone. Bank will never ask for this code.",
			"456012",
		},
		{
			"Code with HTML formatting",
			"Your verification code",
			"<p>Your verification code is:</p><div style='font-size: 24px; font-weight: bold;'>789345</div>",
			"789345",
		},
		{
			"Email with code in HTML tags",
			"Verify your account",
			"<html><body><p>Your code: <span class='otp'>012678</span></p></body></html>",
			"012678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nylas.ExtractOTP(tt.subject, tt.body)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractOTPFalsePositives(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		body    string
	}{
		{
			"Order confirmation email",
			"Your order #12345678 has been confirmed",
			"Thank you for your order #12345678. Your package will arrive in 3-5 business days.",
		},
		{
			"Newsletter with dates",
			"Weekly Newsletter - December 2024",
			"This week in December 2024: Top 10 products of the year. Call us at 555-123-4567.",
		},
		{
			"Invoice email",
			"Invoice #INV-2024-0001",
			"Invoice Amount: $1234.56. Reference: 20241215. Please pay within 30 days.",
		},
		{
			"Shipping notification",
			"Your package is on the way",
			"Tracking number: 1Z999AA10123456784. Expected delivery: December 20, 2024.",
		},
		{
			"Meeting invitation",
			"Meeting at 10:00 AM",
			"Join us for a meeting at 10:00 AM in Conference Room 2024.",
		},
		{
			"Support ticket",
			"Support Ticket #123456",
			"Your support ticket #123456 has been received. Our team will respond within 24 hours.",
		},
		{
			"Flight confirmation",
			"Your flight is booked",
			"Confirmation: XYZ123. Flight AA1234. Departing Dec 25, 2024 at 14:30.",
		},
		{
			"Bank statement summary",
			"Your monthly statement",
			"Account ending in 1234. Balance: $5678.90. Transactions: 12.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nylas.ExtractOTP(tt.subject, tt.body)
			assert.Empty(t, got, "Expected no OTP to be extracted")
		})
	}
}

func TestFindOTP(t *testing.T) {
	t.Run("finds OTP in messages", func(t *testing.T) {
		messages := []domain.Message{
			{
				ID:      "msg-1",
				Subject: "Your verification code",
				Body:    "Your OTP is 123456",
				From:    []domain.EmailParticipant{{Email: "no-reply@example.com"}},
				Date:    time.Now(),
			},
		}

		result, err := nylas.FindOTP(messages)
		assert.NoError(t, err)
		assert.Equal(t, "123456", result.Code)
		assert.Equal(t, "no-reply@example.com", result.From)
	})

	t.Run("returns error when no OTP found", func(t *testing.T) {
		messages := []domain.Message{
			{
				ID:      "msg-1",
				Subject: "Hello",
				Body:    "How are you?",
				From:    []domain.EmailParticipant{{Email: "friend@example.com"}},
				Date:    time.Now(),
			},
		}

		_, err := nylas.FindOTP(messages)
		assert.ErrorIs(t, err, domain.ErrOTPNotFound)
	})

	t.Run("returns error for empty messages", func(t *testing.T) {
		_, err := nylas.FindOTP([]domain.Message{})
		assert.ErrorIs(t, err, domain.ErrOTPNotFound)
	})
}
