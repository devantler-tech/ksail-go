package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"bytes"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

//nolint:dupl // Test structure intentionally mirrors down_test for consistency
func TestHandleUpRunE(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "up")
		seedValidClusterConfig(manager)

		err := HandleUpRunE(cmd, manager, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertOutputContains(t, output.String(), "Cluster created and started successfully")
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		cmd, manager, _ := newCommandAndManager(t, "up")
		manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))

		err := HandleUpRunE(cmd, manager, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if !errors.Is(err, helpers.ErrConfigurationValidationFailed) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})
}

func newCommandAndManager(
	t *testing.T,
	use string,
) (*cobra.Command, *configmanager.ConfigManager, *bytes.Buffer) {
	t.Helper()

	buffer := &bytes.Buffer{}
	cmd := &cobra.Command{Use: use}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)

	manager := configmanager.NewConfigManager(buffer)

	return cmd, manager, buffer
}

func seedValidClusterConfig(manager *configmanager.ConfigManager) {
	manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))
	manager.Viper.Set("spec.distributionConfig", "kind.yaml")
	manager.Viper.Set("spec.connection.context", "")
	manager.Viper.Set("spec.connection.kubeconfig", "~/.kube/config")
}
