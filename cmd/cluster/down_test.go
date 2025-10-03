package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"errors"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
)

// TestHandleDownRunE exercises the success and validation error paths for the down command.

func TestHandleDownRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "down")
		testutils.SeedValidClusterConfig(manager)

		err := HandleDownRunE(cmd, manager, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertOutputContains(
			t,
			output.String(),
			"Cluster stopped and deleted successfully (stub implementation)",
		)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		cmd, manager, _ := testutils.NewCommandAndManager(t, "down")
		manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))

		err := HandleDownRunE(cmd, manager, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if !errors.Is(err, helpers.ErrConfigurationValidationFailed) {
			t.Fatalf("expected validation error, got %v", err)
		}

		if !strings.Contains(err.Error(), "failed to provision cluster down") {
			t.Fatalf("expected contextual error message, got %v", err)
		}

		if !strings.Contains(err.Error(), "failed to load cluster configuration") {
			t.Fatalf("expected wrapped load error, got %v", err)
		}
	})
}
