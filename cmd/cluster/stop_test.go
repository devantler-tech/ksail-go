package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestStopCommandConfigLoad exercises the success and validation error paths for the stop command.
func TestStopCommandConfigLoad(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(t, func() *cobra.Command {
			return NewStopCmd(newTestRuntime())
		})
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(t, func() *cobra.Command {
			return NewStopCmd(newTestRuntime())
		})
	})
}
