package testutils

import (
	"testing"

	"github.com/spf13/cobra"
)

// SetFlags sets multiple flags on a Cobra command, failing the test on first error.
// Keeps tests concise and ensures consistent error handling.
func SetFlags(t *testing.T, cmd *cobra.Command, values map[string]string) {
	t.Helper()
	for k, v := range values {
		if err := cmd.Flags().Set(k, v); err != nil {
			t.Fatalf("failed to set flag %s: %v", k, err)
		}
	}
}
