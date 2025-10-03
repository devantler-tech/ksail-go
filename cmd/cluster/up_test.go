package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
)

// TestHandleUpRunE exercises success and validation error paths.

func TestHandleUpRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "up")
		testutils.SeedValidClusterConfig(manager)

		err := HandleUpRunE(cmd, manager, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertOutputContains(t, output.String(), "Cluster created and started successfully")
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		testutils.RunValidationErrorTest(t, "up", HandleUpRunE)
	})
}
