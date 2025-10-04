package cluster //nolint:testpackage // Access internal helpers without exporting them.

import "testing"

// TestHandleDownRunE exercises the success and validation error paths for the down command.
func TestHandleDownRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(
			t,
			"down",
			HandleDownRunE,
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(
			t,
			"down",
			HandleDownRunE,
			"failed to load cluster configuration",
		)
	})
}
