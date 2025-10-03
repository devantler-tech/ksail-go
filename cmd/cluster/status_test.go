package cluster //nolint:testpackage // Requires internal access to helper functions.

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
)

// TestHandleStatusRunE exercises success and load-failure paths.

func TestHandleStatusRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "status")
		testutils.SeedValidClusterConfig(manager)
		manager.Viper.Set("spec.connection.context", "kind-kind")

		err := HandleStatusRunE(cmd, manager, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertOutputContains(t, output.String(), "Cluster status: Running (stub implementation)")
		// Config loading messages are now printed by the config manager
		assertOutputContains(t, output.String(), "config loaded")
	})

	t.Run("load failure", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		testutils.RunValidationErrorTest(t, "status", HandleStatusRunE)
	})
}

func TestNewStatusCmdIncludesTimeoutSelector(t *testing.T) {
	t.Parallel()

	cmd := NewStatusCmd()

	_, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		t.Fatalf("expected timeout flag to be registered, got error %v", err)
	}
}

// TestHandleStatusRunE_Success validates the happy path behavior.
func TestHandleStatusRunE_Success(t *testing.T) {
	t.Parallel()

	cmd, manager, output := testutils.NewCommandAndManager(t, "status")
	testutils.SeedValidClusterConfig(manager)
	manager.Viper.Set("spec.connection.context", "kind-kind")

	err := HandleStatusRunE(cmd, manager, nil)
	if err != nil {
		// t.Fatalf ensures test stops immediately with context
		t.Fatalf("expected no error, got %v", err)
	}

	assertOutputContains(t, output.String(), "Cluster status: Running (stub implementation)")
	// Config loading messages are now printed by the config manager
	assertOutputContains(t, output.String(), "config loaded")
}

// TestHandleStatusRunE_LoadFailure validates error path when config cannot be loaded.
func TestHandleStatusRunE_LoadFailure(t *testing.T) { //nolint:paralleltest // uses t.Chdir
	testutils.RunValidationErrorTest(t, "status", HandleStatusRunE)
}
