package common

import "github.com/fatih/color"

// Common color definitions used across CLI commands.
// Import these instead of defining package-local color vars.
var (
	Bold      = color.New(color.Bold)
	BoldWhite = color.New(color.FgWhite, color.Bold)
	Dim       = color.New(color.Faint)
	Cyan      = color.New(color.FgCyan)
	Green     = color.New(color.FgGreen)
	Yellow    = color.New(color.FgYellow)
	Red       = color.New(color.FgRed)
	Blue      = color.New(color.FgBlue)
)
