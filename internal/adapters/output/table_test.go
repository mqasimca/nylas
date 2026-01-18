package output

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestTableWriter_WriteList(t *testing.T) {
	type item struct {
		ID    string
		Name  string
		Count int
	}

	tests := []struct {
		name     string
		data     []item
		columns  []ports.Column
		colored  bool
		contains []string
	}{
		{
			name: "basic table",
			data: []item{
				{ID: "1", Name: "Alice", Count: 10},
				{ID: "2", Name: "Bob", Count: 20},
			},
			columns: []ports.Column{
				{Header: "ID", Field: "ID"},
				{Header: "Name", Field: "Name"},
				{Header: "Count", Field: "Count"},
			},
			colored:  false,
			contains: []string{"ID", "NAME", "COUNT", "Alice", "Bob", "10", "20"},
		},
		{
			name:     "empty list",
			data:     []item{},
			columns:  []ports.Column{{Header: "ID", Field: "ID"}},
			colored:  false,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := NewTableWriter(&buf, tt.colored)
			err := tw.WriteList(tt.data, tt.columns)
			assert.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestTableWriter_Write(t *testing.T) {
	type item struct {
		ID   string
		Name string
	}

	var buf bytes.Buffer
	tw := NewTableWriter(&buf, false)
	err := tw.Write(item{ID: "123", Name: "Test"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID:")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "Test")
}

func TestTableWriter_WriteError(t *testing.T) {
	tests := []struct {
		name     string
		colored  bool
		contains string
	}{
		{
			name:     "no color",
			colored:  false,
			contains: "Error: test error",
		},
		{
			name:     "with color",
			colored:  true,
			contains: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := NewTableWriter(&buf, tt.colored)
			testErr := errors.New("test error")
			err := tw.WriteError(testErr)
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"short", 5, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"longer than max", 10, "longer ..."},
		{"ab", 2, "ab"},
		{"abc", 2, "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		expected string
	}{
		{time.Time{}, ""},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-1 * time.Minute), "1 minute ago"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-1 * time.Hour), "1 hour ago"},
		{now.Add(-3 * time.Hour), "3 hours ago"},
		{now.Add(-24 * time.Hour), "yesterday"},
		{now.Add(-72 * time.Hour), "3 days ago"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	type item struct {
		ID   string
		Name string
	}

	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"exact match", "ID", "123"},
		{"case insensitive", "id", "123"},
		{"name field", "Name", "Test"},
		{"non-existent", "Missing", ""},
	}

	v := item{ID: "123", Name: "Test"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldValue(reflect.ValueOf(v), tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"bool true", true, "Yes"},
		{"bool false", false, "No"},
		{"string slice", []string{"a", "b"}, "a, b"},
		{"empty slice", []string{}, ""},
		{"int", 42, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
