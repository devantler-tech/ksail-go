package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
)

//nolint:dupl // Test structure intentionally mirrors up_test for consistency
func TestHandleDownRunE(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "down")
		seedValidClusterConfig(manager)

		err := HandleDownRunE(cmd, manager, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertOutputContains(t, output.String(), "Cluster destroyed successfully")
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		cmd, manager, _ := newCommandAndManager(t, "down")
		manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))

		err := HandleDownRunE(cmd, manager, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if !errors.Is(err, helpers.ErrConfigurationValidationFailed) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})
}
