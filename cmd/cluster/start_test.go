package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestStartCommandConfigLoad exercises the success and validation error paths for the start command.
func TestStartCommandConfigLoad(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(t, func() *cobra.Command {
			return NewStartCmd(newTestRuntime())
		})
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(t, func() *cobra.Command {
			return NewStartCmd(newTestRuntime())
		})
	})
}
