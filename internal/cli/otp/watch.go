package otp

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

func newWatchCmd() *cobra.Command {
	var (
		interval int
		noCopy   bool
	)

	cmd := &cobra.Command{
		Use:   "watch [email]",
		Short: "Watch for new OTP codes",
		Long: `Continuously watch for new OTP codes.

Press Ctrl+C to stop watching.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			otpSvc, err := createOTPService()
			if err != nil {
				return err
			}

			var email string
			if len(args) > 0 {
				email = args[0]
			}

			cyan := common.Cyan
			green := common.Green
			dim := common.Dim

			fmt.Printf("Watching for OTP codes")
			if email != "" {
				fmt.Printf(" for %s", email)
			}
			fmt.Printf(" (every %ds)...\n", interval)
			_, _ = dim.Println("Press Ctrl+C to stop")
			fmt.Println()

			var lastCode string
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			// Check immediately
			checkOTP := func() {
				ctx, cancel := common.CreateContext()
				defer cancel()

				var result *domain.OTPResult
				var err error
				if email != "" {
					result, err = otpSvc.GetOTP(ctx, email)
				} else {
					result, err = otpSvc.GetOTPDefault(ctx)
				}

				if err != nil {
					if err != domain.ErrOTPNotFound {
						_, _ = dim.Printf("[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
					}
					return
				}

				if result.Code != lastCode {
					lastCode = result.Code

					if !noCopy {
						_ = clipboard.WriteAll(result.Code)
					}

					_, _ = cyan.Printf("\n[%s] ", time.Now().Format("15:04:05"))
					_, _ = green.Printf("New OTP: %s\n", result.Code)
					_, _ = dim.Printf("         From: %s\n", result.From)
					_, _ = dim.Printf("         Subject: %s\n", result.Subject)
					if !noCopy {
						_, _ = green.Println("         ✓ Copied to clipboard")
					}
				}
			}

			checkOTP()

			for range ticker.C {
				checkOTP()
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&interval, "interval", "i", 10, "Check interval in seconds")
	cmd.Flags().BoolVar(&noCopy, "no-copy", false, "Don't copy OTP to clipboard")

	return cmd
}
