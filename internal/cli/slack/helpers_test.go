//go:build !integration

package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsChannelID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid channel IDs
		{
			name:     "public channel ID",
			input:    "C01234567890",
			expected: true,
		},
		{
			name:     "private channel ID",
			input:    "G01234567890",
			expected: true,
		},
		{
			name:     "DM channel ID",
			input:    "D01234567890",
			expected: true,
		},
		{
			name:     "minimum valid length C",
			input:    "C12345678",
			expected: true,
		},
		{
			name:     "minimum valid length G",
			input:    "G12345678",
			expected: true,
		},
		{
			name:     "minimum valid length D",
			input:    "D12345678",
			expected: true,
		},

		// Invalid channel IDs
		{
			name:     "channel name without hash",
			input:    "general",
			expected: false,
		},
		{
			name:     "channel name with hash",
			input:    "#general",
			expected: false,
		},
		{
			name:     "too short",
			input:    "C1234567",
			expected: false,
		},
		{
			name:     "invalid prefix A",
			input:    "A01234567890",
			expected: false,
		},
		{
			name:     "invalid prefix lowercase c",
			input:    "c01234567890",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "single character",
			input:    "C",
			expected: false,
		},
		{
			name:     "just numbers",
			input:    "01234567890",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isChannelID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
