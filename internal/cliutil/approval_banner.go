package cliutil

import (
	"fmt"
	"os"
	"strings"
)

// PrintApprovalBanner prints a fixed-width CLI notice for pending device approval.
func PrintApprovalBanner(approvalURL string) {
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "  +----------------------------------------------------------------+")
	fmt.Fprintln(os.Stdout, "  | Device pending admin approval                                  |")
	fmt.Fprintln(os.Stdout, "  | Reporting is paused until an admin approves this device.        |")
	fmt.Fprintln(os.Stdout, "  +----------------------------------------------------------------+")
	fmt.Fprintln(os.Stdout, "  Approval URL:")
	u := strings.TrimSpace(approvalURL)
	if u == "" {
		fmt.Fprintln(os.Stdout, "  (empty)")
	} else {
		// Single line so query strings (e.g. &hash=…) are not split mid-value.
		fmt.Fprintf(os.Stdout, "  %s\n", u)
	}
	fmt.Fprintln(os.Stdout, "")
}
