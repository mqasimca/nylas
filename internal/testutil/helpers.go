// Package testutil provides test utilities.
package testutil

import (
	"github.com/mqasimca/nylas/internal/domain"
)

// TestConfig returns a test configuration.
func TestConfig() *domain.Config {
	return &domain.Config{
		Region:          "us",
		CallbackPort:    8080,
		CopyToClipboard: true,
		WatchInterval:   10,
		Grants: []domain.GrantInfo{
			{
				ID:       "test-grant-id",
				Email:    "test@example.com",
				Provider: domain.ProviderGoogle,
			},
		},
		DefaultGrant: "test-grant-id",
	}
}

// TestGrant returns a test grant.
func TestGrant() *domain.Grant {
	return &domain.Grant{
		ID:          "test-grant-id",
		Email:       "test@example.com",
		Provider:    domain.ProviderGoogle,
		GrantStatus: "valid",
	}
}

// TestGrantInfo returns a test grant info.
func TestGrantInfo() domain.GrantInfo {
	return domain.GrantInfo{
		ID:       "test-grant-id",
		Email:    "test@example.com",
		Provider: domain.ProviderGoogle,
	}
}
