//go:build integration

package integration

import (
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

func TestCLI_AIHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("ai", "--help")

	if err != nil {
		t.Fatalf("ai --help failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"Manage AI/LLM settings",
		"config",
		"AI configuration",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
		}
	}
}

func TestCLI_AIConfigHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("ai", "config", "--help")

	if err != nil {
		t.Fatalf("ai config --help failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"AI configuration",
		"show",
		"list",
		"get",
		"set",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
		}
	}
}
