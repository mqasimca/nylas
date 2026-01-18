package otp

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

func newGetCmd() *cobra.Command {
	var (
		noCopy bool
		raw    bool
	)

	cmd := &cobra.Command{
		Use:   "get [email]",
		Short: "Get the latest OTP code",
		Long: `Get the latest OTP code from your email.

If no email is specified, uses the default account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			otpSvc, err := createOTPService()
			if err != nil {
				return err
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			var result *domain.OTPResult
			if len(args) > 0 {
				result, err = otpSvc.GetOTP(ctx, args[0])
			} else {
				result, err = otpSvc.GetOTPDefault(ctx)
			}

			if err != nil {
				if err == domain.ErrOTPNotFound {
					return fmt.Errorf("no OTP found in recent messages")
				}
				return err
			}

			// Copy to clipboard unless disabled
			if !noCopy {
				_ = clipboard.WriteAll(result.Code)
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				return common.PrintJSON(result)
			}

			if raw {
				fmt.Println(result.Code)
				return nil
			}

			// Display fancy OTP box
			displayOTP(result, !noCopy)

			return nil
		},
	}

	cmd.Flags().BoolVar(&noCopy, "no-copy", false, "Don't copy OTP to clipboard")
	cmd.Flags().BoolVar(&raw, "raw", false, "Output only the OTP code")

	return cmd
}

func displayOTP(result *domain.OTPResult, copied bool) {
	cyan := common.BoldCyan
	green := common.Green
	dim := common.Dim

	// Format code with spaces between digits
	spaced := strings.Join(strings.Split(result.Code, ""), "  ")

	// Draw box
	boxWidth := len(spaced) + 6
	border := strings.Repeat("═", boxWidth)

	_, _ = cyan.Printf("╔%s╗\n", border)
	_, _ = cyan.Printf("║%s║\n", strings.Repeat(" ", boxWidth))
	_, _ = cyan.Printf("║   %s   ║\n", spaced)
	_, _ = cyan.Printf("║%s║\n", strings.Repeat(" ", boxWidth))
	_, _ = cyan.Printf("╚%s╝\n", border)

	fmt.Println()
	_, _ = dim.Printf("From:        %s\n", result.From)
	_, _ = dim.Printf("Subject:     %s\n", result.Subject)
	_, _ = dim.Printf("Received:    %s\n", common.FormatTimeAgo(result.Received))

	if copied {
		_, _ = green.Println("\n✓ Copied to clipboard")
	}
}
