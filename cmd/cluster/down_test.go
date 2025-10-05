package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
)

// TestDownCommandConfigLoad exercises the success and validation error paths for the down command.
func TestDownCommandConfigLoad(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		runLifecycleSuccessCase(
			t,
			"down",
			helpers.HandleConfigLoadRunE,
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		runLifecycleValidationErrorCase(
			t,
			"down",
			helpers.HandleConfigLoadRunE,
			"failed to load cluster configuration",
		)
	})
}
