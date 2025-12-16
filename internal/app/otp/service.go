// Package otp provides OTP-related business logic.
package otp

import (
	"context"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// Service handles OTP operations.
type Service struct {
	client     ports.NylasClient
	grantStore ports.GrantStore
	config     ports.ConfigStore
}

// NewService creates a new OTP service.
func NewService(client ports.NylasClient, grantStore ports.GrantStore, config ports.ConfigStore) *Service {
	return &Service{
		client:     client,
		grantStore: grantStore,
		config:     config,
	}
}

// GetOTP retrieves the most recent OTP for an account.
func (s *Service) GetOTP(ctx context.Context, email string) (*domain.OTPResult, error) {
	// Get grant ID for email
	grant, err := s.grantStore.GetGrantByEmail(email)
	if err != nil {
		return nil, err
	}

	return s.GetOTPByGrantID(ctx, grant.ID)
}

// GetOTPByGrantID retrieves the most recent OTP for a grant.
func (s *Service) GetOTPByGrantID(ctx context.Context, grantID string) (*domain.OTPResult, error) {
	// Fetch recent messages
	messages, err := s.client.GetMessages(ctx, grantID, 20)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, domain.ErrNoMessages
	}

	// Find OTP in messages
	return nylas.FindOTP(messages)
}

// GetOTPDefault retrieves the most recent OTP for the default account.
func (s *Service) GetOTPDefault(ctx context.Context) (*domain.OTPResult, error) {
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil {
		return nil, err
	}

	return s.GetOTPByGrantID(ctx, grantID)
}

// GetMessages retrieves recent messages for an account.
func (s *Service) GetMessages(ctx context.Context, email string, limit int) ([]domain.Message, error) {
	grant, err := s.grantStore.GetGrantByEmail(email)
	if err != nil {
		return nil, err
	}

	return s.client.GetMessages(ctx, grant.ID, limit)
}

// GetMessagesDefault retrieves recent messages for the default account.
func (s *Service) GetMessagesDefault(ctx context.Context, limit int) ([]domain.Message, error) {
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil {
		return nil, err
	}

	return s.client.GetMessages(ctx, grantID, limit)
}

// ListAccounts returns all configured accounts.
func (s *Service) ListAccounts() ([]domain.GrantInfo, error) {
	return s.grantStore.ListGrants()
}
