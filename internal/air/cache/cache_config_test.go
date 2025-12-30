package cache

import (
	"testing"
)

// ================================
// ENCRYPTION HELPER TESTS
// ================================

func TestTableNames(t *testing.T) {
	t.Parallel()

	names := tableNames()

	if len(names) == 0 {
		t.Error("expected non-empty table names")
	}

	// Check that expected tables are present
	expectedTables := map[string]bool{
		"emails":   false,
		"events":   false,
		"contacts": false,
		"folders":  false,
	}

	for _, name := range names {
		if _, ok := expectedTables[name]; ok {
			expectedTables[name] = true
		}
	}

	for table, found := range expectedTables {
		if !found {
			t.Errorf("expected table '%s' in tableNames()", table)
		}
	}
}

func TestAllowedTables(t *testing.T) {
	t.Parallel()

	// Test that the allowedTables map is populated
	if len(allowedTables) == 0 {
		t.Error("allowedTables should not be empty")
	}

	// Check specific tables
	if !allowedTables["emails"] {
		t.Error("emails should be in allowedTables")
	}
	if !allowedTables["events"] {
		t.Error("events should be in allowedTables")
	}
	if !allowedTables["contacts"] {
		t.Error("contacts should be in allowedTables")
	}

	// Check that random strings are not allowed
	if allowedTables["malicious_table"] {
		t.Error("malicious_table should not be in allowedTables")
	}
	if allowedTables["DROP TABLE emails"] {
		t.Error("SQL injection attempt should not be in allowedTables")
	}
}

func TestEncryptionConfig_Struct(t *testing.T) {
	t.Parallel()

	config := EncryptionConfig{
		Enabled: true,
		KeyID:   "user@example.com",
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
	if config.KeyID != "user@example.com" {
		t.Errorf("expected KeyID 'user@example.com', got '%s'", config.KeyID)
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	if config.BasePath == "" {
		t.Error("expected BasePath to be set")
	}
	if config.TTLDays <= 0 {
		t.Error("expected TTLDays to be positive")
	}
	if config.MaxSizeMB <= 0 {
		t.Error("expected MaxSizeMB to be positive")
	}
	if config.SyncIntervalMinutes <= 0 {
		t.Error("expected SyncIntervalMinutes to be positive")
	}
}

// ================================
// ADDITIONAL EMAIL STORE TESTS
// ================================
