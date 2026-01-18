//go:build !integration

package common

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat_AllFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OutputFormat
		hasError bool
	}{
		// Table format variants
		{"table lowercase", "table", FormatTable, false},
		{"table uppercase", "TABLE", FormatTable, false},
		{"table mixed case", "Table", FormatTable, false},
		{"empty defaults to table", "", FormatTable, false},

		// JSON format variants
		{"json lowercase", "json", FormatJSON, false},
		{"json uppercase", "JSON", FormatJSON, false},
		{"json mixed case", "Json", FormatJSON, false},

		// CSV format variants
		{"csv lowercase", "csv", FormatCSV, false},
		{"csv uppercase", "CSV", FormatCSV, false},
		{"csv mixed case", "Csv", FormatCSV, false},

		// YAML format variants
		{"yaml lowercase", "yaml", FormatYAML, false},
		{"yaml uppercase", "YAML", FormatYAML, false},
		{"yml shorthand", "yml", FormatYAML, false},
		{"YML uppercase", "YML", FormatYAML, false},

		// Invalid formats
		{"invalid format", "invalid", "", true},
		{"xml not supported", "xml", "", true},
		{"html not supported", "html", "", true},
		{"spaces not trimmed", " json ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := ParseFormat(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid format")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, format)
			}
		})
	}
}

func TestFormatter_JSON_Output(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		contains []string
	}{
		{
			name:     "simple map",
			data:     map[string]string{"key": "value"},
			contains: []string{`"key"`, `"value"`},
		},
		{
			name:     "slice of maps",
			data:     []map[string]int{{"a": 1}, {"b": 2}},
			contains: []string{`"a"`, `"b"`, "1", "2"},
		},
		{
			name: "struct",
			data: struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{Name: "test", Count: 42},
			contains: []string{`"name"`, `"test"`, `"count"`, "42"},
		},
		{
			name:     "nested struct",
			data:     map[string]any{"outer": map[string]string{"inner": "value"}},
			contains: []string{`"outer"`, `"inner"`, `"value"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatJSON).SetWriter(&buf)

			err := formatter.Format(tt.data)
			require.NoError(t, err)

			output := buf.String()
			for _, s := range tt.contains {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestFormatter_YAML_Output(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		contains []string
	}{
		{
			name:     "simple map",
			data:     map[string]string{"key": "value"},
			contains: []string{"key:", "value"},
		},
		{
			name:     "multiple fields",
			data:     map[string]int{"count": 10, "total": 100},
			contains: []string{"count:", "10", "total:", "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatYAML).SetWriter(&buf)

			err := formatter.Format(tt.data)
			require.NoError(t, err)

			output := buf.String()
			for _, s := range tt.contains {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestFormatter_CSV_Slice(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
		Tag   string `json:"tag"`
	}

	tests := []struct {
		name     string
		data     []Item
		contains []string
	}{
		{
			name: "multiple items",
			data: []Item{
				{Name: "item1", Value: 1, Tag: "a"},
				{Name: "item2", Value: 2, Tag: "b"},
			},
			contains: []string{"name", "value", "tag", "item1", "item2", "1", "2", "a", "b"},
		},
		{
			name:     "single item",
			data:     []Item{{Name: "only", Value: 99, Tag: "x"}},
			contains: []string{"name", "value", "only", "99", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatCSV).SetWriter(&buf)

			err := formatter.Format(tt.data)
			require.NoError(t, err)

			output := buf.String()
			for _, s := range tt.contains {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestFormatter_CSV_EmptySlice(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}

	var buf bytes.Buffer
	formatter := NewFormatter(FormatCSV).SetWriter(&buf)

	err := formatter.Format([]Item{})
	require.NoError(t, err)

	// Empty slice should produce no output
	assert.Empty(t, buf.String())
}

func TestFormatter_CSV_SingleItem(t *testing.T) {
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var buf bytes.Buffer
	formatter := NewFormatter(FormatCSV).SetWriter(&buf)

	// Test single item (not in slice)
	err := formatter.Format(Item{ID: "123", Name: "test"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "id")
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "test")
}

func TestFormatter_CSV_NonStructTypes(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter(FormatCSV).SetWriter(&buf)

	// Non-struct types should fall back to "value" header
	err := formatter.Format("simple string")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "value")
	assert.Contains(t, output, "simple string")
}

func TestGetCSVHeaders(t *testing.T) {
	type TestStruct struct {
		Public     string `json:"public_field"`
		NoTag      string
		SkipField  string `json:"-"`
		unexported string //nolint:unused
	}

	tests := []struct {
		name     string
		data     any
		expected []string
	}{
		{
			name:     "struct with json tags",
			data:     TestStruct{Public: "val", NoTag: "val2"},
			expected: []string{"public_field", "NoTag"},
		},
		{
			name:     "non-struct returns value",
			data:     "string",
			expected: []string{"value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to use reflection to test this internal function
			// Test through Format instead
			var buf bytes.Buffer
			formatter := NewFormatter(FormatCSV).SetWriter(&buf)

			switch v := tt.data.(type) {
			case TestStruct:
				err := formatter.Format(v)
				require.NoError(t, err)
				output := buf.String()
				for _, exp := range tt.expected {
					assert.Contains(t, output, exp)
				}
			case string:
				err := formatter.Format(v)
				require.NoError(t, err)
				output := buf.String()
				assert.Contains(t, output, "value")
			}
		})
	}
}

func TestFormatValue_SpecialTypes(t *testing.T) {
	type ItemWithSlice struct {
		Tags []string `json:"tags"`
	}

	tests := []struct {
		name     string
		data     any
		contains string
	}{
		{
			name:     "slice field",
			data:     []ItemWithSlice{{Tags: []string{"a", "b", "c"}}},
			contains: "a; b; c",
		},
		{
			name:     "empty slice field",
			data:     []ItemWithSlice{{Tags: []string{}}},
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatCSV).SetWriter(&buf)

			err := formatter.Format(tt.data)
			require.NoError(t, err)

			output := buf.String()
			if tt.contains != "" {
				assert.Contains(t, output, tt.contains)
			}
		})
	}
}

func TestTable_BasicOperations(t *testing.T) {
	ResetLogger()
	InitLogger(false, false)

	t.Run("create and render table", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("ID", "NAME", "STATUS").SetWriter(&buf)

		table.AddRow("1", "First", "Active")
		table.AddRow("2", "Second", "Inactive")

		assert.Equal(t, 2, table.RowCount())

		table.Render()
		output := buf.String()

		// Check headers are present
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "NAME")
		assert.Contains(t, output, "STATUS")

		// Check data is present
		assert.Contains(t, output, "First")
		assert.Contains(t, output, "Second")
		assert.Contains(t, output, "Active")
		assert.Contains(t, output, "Inactive")
	})

	t.Run("table with right alignment", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("NAME", "COUNT").SetWriter(&buf)
		table.AlignRight(1) // Align COUNT column to right
		table.AddRow("Items", "100")
		table.Render()

		assert.Contains(t, buf.String(), "100")
	})

	t.Run("table with short rows", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("A", "B", "C").SetWriter(&buf)

		// Add row with fewer values than headers
		table.AddRow("only one")
		table.Render()

		assert.Equal(t, 1, table.RowCount())
		assert.Contains(t, buf.String(), "only one")
	})

	t.Run("table with invalid column index", func(t *testing.T) {
		table := NewTable("A", "B")
		// This should not panic - should silently ignore invalid index
		table.AlignRight(10)
		table.AddRow("a", "b")
		// No assertion needed - just verify no panic
	})
}

func TestTable_QuietMode(t *testing.T) {
	ResetLogger()
	InitLogger(false, true) // Enable quiet mode

	var buf bytes.Buffer
	table := NewTable("HEADER").SetWriter(&buf)
	table.AddRow("value")
	table.Render()

	// In quiet mode, should not produce output
	assert.Empty(t, buf.String())
}

func TestPrintFunctions_QuietMode(t *testing.T) {
	ResetLogger()
	InitLogger(false, true) // Enable quiet mode

	// These should not panic in quiet mode
	PrintSuccess("success: %s", "test")
	PrintWarning("warning: %s", "test")
	PrintInfo("info: %s", "test")
}

func TestPrintError_AlwaysPrints(t *testing.T) {
	ResetLogger()
	InitLogger(false, true) // Enable quiet mode

	// PrintError should print even in quiet mode (to stderr)
	// We can't easily capture stderr, so just verify no panic
	PrintError("error: %s", "test")
}

func TestTable_UTF8Support(t *testing.T) {
	ResetLogger()
	InitLogger(false, false)

	t.Run("handles UTF-8 characters correctly", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("NAME", "EMOJI").SetWriter(&buf)

		// UTF-8 characters should be counted by runes, not bytes
		table.AddRow("Alice", "👋🌍")
		table.AddRow("Bob", "🚀✨")

		table.Render()
		output := buf.String()

		assert.Contains(t, output, "👋🌍")
		assert.Contains(t, output, "🚀✨")
	})

	t.Run("handles mixed ASCII and UTF-8", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("ID", "DESCRIPTION").SetWriter(&buf)

		table.AddRow("1", "Hello 世界")
		table.AddRow("2", "Test data")

		table.Render()
		output := buf.String()

		assert.Contains(t, output, "Hello 世界")
		assert.Contains(t, output, "Test data")
	})
}

func TestTable_MaxWidth(t *testing.T) {
	ResetLogger()
	InitLogger(false, false)

	t.Run("truncates long text with max width", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("ID", "DESCRIPTION").SetWriter(&buf)
		table.SetMaxWidth(1, 20) // Limit DESCRIPTION column to 20 chars

		table.AddRow("1", "This is a very long description that should be truncated")
		table.AddRow("2", "Short")

		table.Render()
		output := buf.String()

		// Should contain truncated version with ellipsis
		assert.Contains(t, output, "...")
		assert.Contains(t, output, "Short")
		assert.NotContains(t, output, "should be truncated")
	})

	t.Run("no truncation when text is under max width", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("NAME").SetWriter(&buf)
		table.SetMaxWidth(0, 50)

		table.AddRow("Short text")

		table.Render()
		output := buf.String()

		assert.Contains(t, output, "Short text")
		assert.NotContains(t, output, "...")
	})

	t.Run("no max width when not set", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("CONTENT").SetWriter(&buf)

		longText := "This is a very long piece of text that should not be truncated"
		table.AddRow(longText)

		table.Render()
		output := buf.String()

		assert.Contains(t, output, longText)
	})
}

func TestTruncateCell(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"truncate with ellipsis", "hello world", 8, "hello..."},
		{"exact length", "exactly", 7, "exactly"},
		{"very short maxLen", "test", 2, "te"},
		{"UTF-8 characters", "Hello 世界!", 7, "Hell..."},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateCell(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			// Verify result doesn't exceed maxLen in rune count
			assert.LessOrEqual(t, len([]rune(result)), tt.maxLen)
		})
	}
}

func TestPadString(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		width      int
		alignRight bool
		expected   string
	}{
		{"pad right with spaces", "hello", 10, false, "hello     "},
		{"pad left with spaces", "world", 10, true, "     world"},
		{"no padding needed", "exactly", 7, false, "exactly"},
		{"UTF-8 left align", "Hi 世界", 8, false, "Hi 世界   "},
		{"UTF-8 right align", "Hi 世界", 8, true, "   Hi 世界"},
		{"emoji left align", "👋🌍", 5, false, "👋🌍   "},
		{"emoji right align", "👋🌍", 5, true, "   👋🌍"},
		{"no padding for longer string", "too long text", 5, false, "too long text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padString(tt.input, tt.width, tt.alignRight)
			assert.Equal(t, tt.expected, result)
			// Verify result width matches expected width (or is longer if input was too long)
			assert.GreaterOrEqual(t, len([]rune(result)), len([]rune(tt.input)))
		})
	}
}

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no ansi codes", "hello world", "hello world"},
		{"simple color code", "\x1b[31mred text\x1b[0m", "red text"},
		{"bold code", "\x1b[1mbold\x1b[0m", "bold"},
		{"multiple codes", "\x1b[32mgreen\x1b[0m and \x1b[34mblue\x1b[0m", "green and blue"},
		{"empty string", "", ""},
		{"only ansi codes", "\x1b[31m\x1b[0m", ""},
		{"cyan sprint", Cyan.Sprint("test"), "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripAnsiCodes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVisualWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"plain text", "hello", 5},
		{"with ansi codes", "\x1b[31mred\x1b[0m", 3},
		{"UTF-8 chars", "世界", 2},
		{"UTF-8 with ansi", "\x1b[32m世界\x1b[0m", 2},
		{"emoji", "👋🌍", 2},
		{"cyan sprint", Cyan.Sprint("test"), 4},
		{"green sprint Yes", Green.Sprint("Yes"), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := visualWidth(tt.input)
			assert.Equal(t, tt.expected, result, "visualWidth(%q) should be %d", tt.input, tt.expected)
		})
	}
}

func TestTable_Alignment(t *testing.T) {
	ResetLogger()
	InitLogger(false, false)

	t.Run("columns align properly with varying widths", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("SHORT", "MEDIUM LENGTH", "L").SetWriter(&buf)

		table.AddRow("A", "B", "C")
		table.AddRow("Long data", "X", "Very long content here")

		table.Render()
		output := buf.String()

		// Just verify all expected content is present and properly formatted
		assert.Contains(t, output, "SHORT")
		assert.Contains(t, output, "MEDIUM LENGTH")
		assert.Contains(t, output, "Long data")
		assert.Contains(t, output, "Very long content here")

		// Verify we have multiple lines (header + separator + rows)
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.GreaterOrEqual(t, len(lines), 4, "Should have header, separator, and data rows")
	})

	t.Run("UTF-8 characters align correctly", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("NAME", "EMOJI", "COUNT").SetWriter(&buf)

		table.AddRow("Alice", "👋", "1")
		table.AddRow("Bob", "🌍🚀", "2")
		table.AddRow("测试", "✨", "3")

		table.Render()
		output := buf.String()

		// Verify all expected content is present
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "👋")
		assert.Contains(t, output, "🌍🚀")
		assert.Contains(t, output, "测试")
		assert.Contains(t, output, "✨")
	})

	t.Run("right alignment works correctly", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("NAME", "AMOUNT").SetWriter(&buf)
		table.AlignRight(1) // Right-align the AMOUNT column

		table.AddRow("Item A", "100")
		table.AddRow("Item B", "50")
		table.AddRow("Item C", "1234")

		table.Render()
		output := buf.String()

		// Output should contain all values
		assert.Contains(t, output, "Item A")
		assert.Contains(t, output, "100")
		assert.Contains(t, output, "1234")
	})
}
