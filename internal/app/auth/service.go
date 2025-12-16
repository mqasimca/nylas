// Package auth provides authentication-related business logic.
package auth

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// Service handles authentication operations.
type Service struct {
	client     ports.NylasClient
	grantStore ports.GrantStore
	config     ports.ConfigStore
	server     ports.OAuthServer
	browser    ports.Browser
}

// NewService creates a new auth service.
func NewService(
	client ports.NylasClient,
	grantStore ports.GrantStore,
	config ports.ConfigStore,
	server ports.OAuthServer,
	browser ports.Browser,
) *Service {
	return &Service{
		client:     client,
		grantStore: grantStore,
		config:     config,
		server:     server,
		browser:    browser,
	}
}

// Login performs OAuth login with the specified provider.
func (s *Service) Login(ctx context.Context, provider domain.Provider) (*domain.Grant, error) {
	// Start callback server
	if err := s.server.Start(); err != nil {
		return nil, err
	}
	defer s.server.Stop()

	// Build auth URL and open browser
	authURL := s.client.BuildAuthURL(provider, s.server.GetRedirectURI())
	if err := s.browser.Open(authURL); err != nil {
		return nil, err
	}

	// Wait for callback
	code, err := s.server.WaitForCallback(ctx)
	if err != nil {
		return nil, err
	}

	// Exchange code for tokens
	grant, err := s.client.ExchangeCode(ctx, code, s.server.GetRedirectURI())
	if err != nil {
		return nil, err
	}

	// Save grant info
	grantInfo := domain.GrantInfo{
		ID:       grant.ID,
		Email:    grant.Email,
		Provider: grant.Provider,
	}
	if err := s.grantStore.SaveGrant(grantInfo); err != nil {
		return nil, err
	}

	// Set as default if no default exists or this is the first grant
	isFirstGrant := false
	if _, err := s.grantStore.GetDefaultGrant(); err == domain.ErrNoDefaultGrant {
		_ = s.grantStore.SetDefaultGrant(grant.ID)
		isFirstGrant = true
	}

	// Update config with grant (only update default if this is the first grant)
	cfg, _ := s.config.Load()
	cfg.Grants = append(cfg.Grants, grantInfo)
	if isFirstGrant {
		cfg.DefaultGrant = grant.ID
	}
	_ = s.config.Save(cfg)

	return grant, nil
}

// Logout revokes the current grant.
func (s *Service) Logout(ctx context.Context) error {
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil {
		return err
	}

	// Revoke on Nylas
	if err := s.client.RevokeGrant(ctx, grantID); err != nil && err != domain.ErrGrantNotFound {
		return err
	}

	// Remove from local storage
	if err := s.grantStore.DeleteGrant(grantID); err != nil {
		return err
	}

	// Auto-switch to another grant if available
	s.autoSwitchDefault()

	return nil
}

// LogoutGrant revokes a specific grant.
func (s *Service) LogoutGrant(ctx context.Context, grantID string) error {
	// Check if this is the default grant
	defaultID, _ := s.grantStore.GetDefaultGrant()
	isDefault := grantID == defaultID

	// Revoke on Nylas
	if err := s.client.RevokeGrant(ctx, grantID); err != nil && err != domain.ErrGrantNotFound {
		return err
	}

	// Remove from local storage
	if err := s.grantStore.DeleteGrant(grantID); err != nil {
		return err
	}

	// Auto-switch to another grant if we deleted the default
	if isDefault {
		s.autoSwitchDefault()
	}

	return nil
}

// autoSwitchDefault sets a new default grant from remaining grants.
func (s *Service) autoSwitchDefault() {
	grants, err := s.grantStore.ListGrants()
	if err != nil || len(grants) == 0 {
		return
	}
	// Set the first remaining grant as default
	_ = s.grantStore.SetDefaultGrant(grants[0].ID)
}
