package testutils

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// RunTestMainWithSnapshotCleanup runs the standard TestMain pattern with snapshot cleanup.
// Shared across packages that only need snapshot cleanup (non command-specific logic).
func RunTestMainWithSnapshotCleanup(m *testing.M) {
	exitCode := m.Run()

	_, err := snaps.Clean(m, snaps.CleanOpts{Sort: true})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to clean snapshots: " + err.Error() + "\n")
		os.Exit(1)
	}

	os.Exit(exitCode)
}
