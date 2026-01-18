// Package ports defines interfaces for the application.
package ports

import "io"

// OutputFormat represents the format for CLI output.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatQuiet OutputFormat = "quiet"
)

// OutputWriter handles formatted output for CLI commands.
type OutputWriter interface {
	// Write outputs a single object (struct, map, etc.)
	Write(data any) error

	// WriteList outputs a list of objects as a table or list format.
	// columns specifies which fields to display and their order.
	WriteList(data any, columns []Column) error

	// WriteError outputs an error message in the appropriate format.
	WriteError(err error) error
}

// Column defines a column for table output.
type Column struct {
	// Header is the column header text
	Header string

	// Field is the struct field name or map key to extract
	Field string

	// Width is the optional fixed width (0 = auto, -1 = no truncation)
	Width int
}

// OutputOptions configures output behavior.
type OutputOptions struct {
	// Format specifies the output format
	Format OutputFormat

	// NoColor disables colored output
	NoColor bool

	// Writer is the destination for output
	Writer io.Writer
}

// QuietFielder can be implemented by types to specify their quiet-mode output.
type QuietFielder interface {
	// QuietField returns the field value to output in quiet mode (usually ID)
	QuietField() string
}
