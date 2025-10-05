package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
)

// TestStopCommandConfigLoad exercises the success and validation error paths for the stop command.
func TestStopCommandConfigLoad(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(
			t,
			"stop",
			helpers.HandleConfigLoadRunE,
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(
			t,
			"stop",
			helpers.HandleConfigLoadRunE,
			"failed to load cluster configuration",
		)
	})
}
