package output

import (
	"fmt"
	"io"

	"github.com/mqasimca/nylas/internal/ports"
	"gopkg.in/yaml.v3"
)

// YAMLWriter outputs data in YAML format.
type YAMLWriter struct {
	w io.Writer
}

// NewYAMLWriter creates a new YAML writer.
func NewYAMLWriter(w io.Writer) *YAMLWriter {
	return &YAMLWriter{w: w}
}

// Write outputs a single object as YAML.
func (yw *YAMLWriter) Write(data any) error {
	return yw.encode(data)
}

// WriteList outputs a list as YAML (columns are ignored).
func (yw *YAMLWriter) WriteList(data any, _ []ports.Column) error {
	return yw.encode(data)
}

// WriteError outputs an error as YAML.
func (yw *YAMLWriter) WriteError(err error) error {
	errObj := map[string]string{
		"error": err.Error(),
	}
	return yw.encode(errObj)
}

// encode writes data as YAML.
func (yw *YAMLWriter) encode(data any) error {
	enc := yaml.NewEncoder(yw.w)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	return nil
}
