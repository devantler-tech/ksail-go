package cluster //nolint:testpackage // Access internal helpers without exporting them.

import "testing"

// TestHandleStopRunE exercises the success and validation error paths for the stop command.

func TestHandleStopRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(
			t,
			"stop",
			HandleStopRunE,
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(
			t,
			"stop",
			HandleStopRunE,
			"failed to provision cluster stop",
			"failed to load cluster configuration",
		)
	})
}
