package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
)

// TestStartCommandConfigLoad exercises the success and validation error paths for the start command.
func TestStartCommandConfigLoad(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(
			t,
			"start",
			helpers.HandleConfigLoadRunE,
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(
			t,
			"start",
			helpers.HandleConfigLoadRunE,
			"failed to load cluster configuration",
		)
	})
}
