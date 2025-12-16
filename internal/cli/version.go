package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "nylas version %s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  Commit:     %s\n", Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  Built:      %s\n", BuildDate)
			fmt.Fprintf(cmd.OutOrStdout(), "  Go version: %s\n", runtime.Version())
			fmt.Fprintf(cmd.OutOrStdout(), "  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
