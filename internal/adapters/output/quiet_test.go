package output

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mqasimca/nylas/internal/ports"
	"github.com/stretchr/testify/assert"
)

type testItem struct {
	ID   string
	Name string
}

type quietItem struct {
	ID   string
	Name string
}

func (q quietItem) QuietField() string {
	return q.ID
}

func TestQuietWriter_Write(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "struct with ID field",
			data:     testItem{ID: "123", Name: "Test"},
			expected: "123\n",
		},
		{
			name:     "implements QuietFielder",
			data:     quietItem{ID: "456", Name: "Test"},
			expected: "456\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			qw := NewQuietWriter(&buf)
			err := qw.Write(tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestQuietWriter_WriteList(t *testing.T) {
	data := []testItem{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
	}
	columns := []ports.Column{{Header: "ID", Field: "ID"}}

	var buf bytes.Buffer
	qw := NewQuietWriter(&buf)
	err := qw.WriteList(data, columns)
	assert.NoError(t, err)

	output := buf.String()
	assert.Equal(t, "1\n2\n", output)
}

func TestQuietWriter_WriteError(t *testing.T) {
	var buf bytes.Buffer
	qw := NewQuietWriter(&buf)
	err := qw.WriteError(errors.New("test error"))
	assert.NoError(t, err)
	// Quiet mode doesn't output errors
	assert.Empty(t, buf.String())
}

func TestExtractQuietField(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "struct with ID",
			data:     testItem{ID: "123"},
			expected: "123",
		},
		{
			name:     "implements QuietFielder",
			data:     quietItem{ID: "456"},
			expected: "456",
		},
		{
			name:     "simple string",
			data:     "hello",
			expected: "hello",
		},
		{
			name:     "int",
			data:     42,
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractQuietField(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}
