package cluster //nolint:testpackage // Requires internal access to helper functions.

import (
	"strings"
	"testing"
)

// TestStatusCommandConfigLoad exercises success and load-failure paths.
func TestStatusCommandConfigLoad(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		output := runLifecycleSuccessCase(t, NewStatusCmd)
		if strings.Contains(output, "stub implementation") {
			t.Fatalf("unexpected stub output for status command: %q", output)
		}
	})

	t.Run("load failure", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(t, NewStatusCmd)
	})
}
