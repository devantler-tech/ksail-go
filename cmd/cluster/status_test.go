package cluster //nolint:testpackage // Requires internal access to helper functions.

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
)

func TestHandleStatusRunE(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "status")
		seedValidClusterConfig(manager)
		manager.Viper.Set("spec.connection.context", "kind-kind")

		err := HandleStatusRunE(cmd, manager, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertOutputContains(t, output.String(), "Cluster status: Running (stub implementation)")
	})

	t.Run("load failure", func(t *testing.T) {
		t.Parallel()

		cmd, manager, _ := newIsolatedCommandAndManager(t, "status")
		// Don't seed config - LoadConfig will use defaults, then validation will fail
		// Set invalid distribution without distributionConfig to trigger validation error
		manager.Viper.Set("metadata.name", "test-cluster")

		err := HandleStatusRunE(cmd, manager, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if !errors.Is(err, helpers.ErrConfigurationValidationFailed) {
			t.Fatalf("expected validation error, got %v", err)
		}
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
