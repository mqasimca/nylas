package output

import (
	"bytes"
	"testing"

	"github.com/mqasimca/nylas/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestNewWriter(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name     string
		format   ports.OutputFormat
		expected string
	}{
		{"table format", ports.FormatTable, "*output.TableWriter"},
		{"json format", ports.FormatJSON, "*output.JSONWriter"},
		{"yaml format", ports.FormatYAML, "*output.YAMLWriter"},
		{"quiet format", ports.FormatQuiet, "*output.QuietWriter"},
		{"empty format defaults to table", "", "*output.TableWriter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ports.OutputOptions{
				Format: tt.format,
				Writer: &buf,
			}
			writer := NewWriter(&buf, opts)
			assert.NotNil(t, writer)
		})
	}
}
