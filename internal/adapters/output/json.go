package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mqasimca/nylas/internal/ports"
)

// JSONWriter outputs data in JSON format.
type JSONWriter struct {
	w io.Writer
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(w io.Writer) *JSONWriter {
	return &JSONWriter{w: w}
}

// Write outputs a single object as JSON.
func (jw *JSONWriter) Write(data any) error {
	return jw.encode(data)
}

// WriteList outputs a list as JSON (columns are ignored).
func (jw *JSONWriter) WriteList(data any, _ []ports.Column) error {
	return jw.encode(data)
}

// WriteError outputs an error as JSON.
func (jw *JSONWriter) WriteError(err error) error {
	errObj := map[string]string{
		"error": err.Error(),
	}
	return jw.encode(errObj)
}

// encode writes data as pretty-printed JSON.
func (jw *JSONWriter) encode(data any) error {
	enc := json.NewEncoder(jw.w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
