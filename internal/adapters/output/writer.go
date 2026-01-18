// Package output provides output formatting adapters for CLI commands.
package output

import (
	"io"

	"github.com/mqasimca/nylas/internal/ports"
)

// NewWriter creates an output writer for the specified format.
func NewWriter(w io.Writer, opts ports.OutputOptions) ports.OutputWriter {
	switch opts.Format {
	case ports.FormatJSON:
		return NewJSONWriter(w)
	case ports.FormatYAML:
		return NewYAMLWriter(w)
	case ports.FormatQuiet:
		return NewQuietWriter(w)
	default:
		return NewTableWriter(w, !opts.NoColor)
	}
}
