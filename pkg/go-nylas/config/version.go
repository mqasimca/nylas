package config

import (
	"fmt"
	"os"
	"strings"
)

// MinimumCLIVersion is the minimum required version of the Nylas CLI.
const MinimumCLIVersion = "2.0.0"

// CheckCompatibility checks if the plugin is compatible with the current CLI version.
// The core CLI sets NYLAS_CLI_VERSION environment variable when executing plugins.
func CheckCompatibility() error {
	cliVersion := os.Getenv("NYLAS_CLI_VERSION")
	if cliVersion == "" {
		return fmt.Errorf("NYLAS_CLI_VERSION not set - this plugin must be run via Nylas CLI")
	}

	// Simple version comparison (can be enhanced with semver library)
	if !isVersionCompatible(cliVersion, MinimumCLIVersion) {
		return fmt.Errorf("plugin requires Nylas CLI v%s or later, but running v%s", MinimumCLIVersion, cliVersion)
	}

	return nil
}

// GetCLIVersion returns the current CLI version from environment.
func GetCLIVersion() string {
	return os.Getenv("NYLAS_CLI_VERSION")
}

// isVersionCompatible checks if current version meets minimum requirement.
// Simplified version comparison - in production, use go-version library.
func isVersionCompatible(current, minimum string) bool {
	current = strings.TrimPrefix(current, "v")
	minimum = strings.TrimPrefix(minimum, "v")

	currentParts := strings.Split(current, ".")
	minimumParts := strings.Split(minimum, ".")

	for i := 0; i < len(minimumParts) && i < len(currentParts); i++ {
		if currentParts[i] > minimumParts[i] {
			return true
		}
		if currentParts[i] < minimumParts[i] {
			return false
		}
	}

	return true
}
