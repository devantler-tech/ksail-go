package testutils

import (
	"os"
	"strings"
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

// ExpectNoError fails the test if err is not nil.
func ExpectNoError(t *testing.T, err error, description string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: unexpected error: %v", description, err)
	}
}

// ExpectErrorContains fails the test if err is nil or does not contain substr.
func ExpectErrorContains(t *testing.T, err error, substr, description string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected error containing %q but got nil", description, substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("%s: expected error to contain %q, got %q", description, substr, err.Error())
	}
}

// ExpectNotNil fails the test if value is nil.
func ExpectNotNil(t *testing.T, value any, description string) {
	t.Helper()

	if value == nil {
		t.Fatalf("expected %s to be non-nil", description)
	}
}

// ExpectTrue fails the test if condition is false.
func ExpectTrue(t *testing.T, condition bool, description string) {
	t.Helper()

	if !condition {
		t.Fatalf("expected %s to be true", description)
	}
}
