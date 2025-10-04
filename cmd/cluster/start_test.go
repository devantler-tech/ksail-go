package cluster //nolint:testpackage // Access internal helpers without exporting them.

import "testing"

// TestHandleStartRunE exercises the success and validation error paths for the start command.

func TestHandleStartRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(
			t,
			"start",
			HandleStartRunE,
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(
			t,
			"start",
			HandleStartRunE,
			"failed to load cluster configuration",
		)
	})
}
