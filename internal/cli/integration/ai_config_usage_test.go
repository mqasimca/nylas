//go:build integration

package integration

import (
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

func TestCLI_AIUsageCommand(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Test usage command (should work even with no data)
	stdout, stderr, err := runCLI("ai", "usage")

	if err != nil {
		t.Fatalf("ai usage failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"AI Usage for",
		"Total Requests:",
		"Ollama:",
		"Claude:",
		"OpenAI:",
		"Groq:",
		"OpenRouter:",
		"Total Tokens:",
		"Estimated Cost:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}
}

func TestCLI_AIUsageJSON(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Test usage command with JSON output
	stdout, stderr, err := runCLI("ai", "usage", "--json")

	if err != nil {
		t.Fatalf("ai usage --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON
	if !strings.Contains(stdout, "{") || !strings.Contains(stdout, "}") {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	// Should contain expected fields
	expectedFields := []string{
		"month",
		"total_requests",
		"estimated_cost",
	}

	for _, field := range expectedFields {
		if !strings.Contains(stdout, field) {
			t.Errorf("Expected JSON to contain field %q\nGot: %s", field, stdout)
		}
	}
}

func TestCLI_AISetBudget(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Set a budget
	stdout, stderr, err := runCLI("ai", "set-budget", "--monthly", "50")

	if err != nil {
		t.Fatalf("ai set-budget failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"✓ Monthly budget set to $50.00",
		"Alert threshold: 80%",
		"Budget applies to:",
		"Claude",
		"OpenAI",
		"Groq",
		"OpenRouter",
		"Ollama (local) usage is free",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}

	// Verify with show-budget
	stdout, stderr, err = runCLI("ai", "show-budget")

	if err != nil {
		t.Fatalf("ai show-budget failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Monthly Limit:       $50.00") {
		t.Errorf("Expected show-budget to show $50.00 limit\nGot: %s", stdout)
	}
}

func TestCLI_AIClearData(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Clear data with --force flag (no confirmation prompt)
	stdout, stderr, err := runCLI("ai", "clear-data", "--force")

	if err != nil {
		t.Fatalf("ai clear-data failed: %v\nstderr: %s", err, stderr)
	}

	// Should succeed (even if no data exists)
	expectedStrings := []string{
		"AI data cleared",
		"Learned scheduling patterns",
		"Usage statistics",
		"Cached responses",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}
}

// Privacy and Features Config Show Test
