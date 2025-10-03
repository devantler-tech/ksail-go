// Package testutils provides shared helpers for cluster command tests.
package testutils

import (
	"bytes"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewCommandAndManager creates a new cobra.Command, config manager, and output buffer for tests.
func NewCommandAndManager(
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

// SeedValidClusterConfig sets up a valid cluster config in the manager.
func SeedValidClusterConfig(manager *configmanager.ConfigManager) {
	manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))
	manager.Viper.Set("spec.distributionConfig", "kind.yaml")
	manager.Viper.Set("spec.connection.context", "")
	manager.Viper.Set("spec.connection.kubeconfig", "~/.kube/config")
}

// RunValidationErrorTest runs a validation error test for a given command handler.
func RunValidationErrorTest(
	t *testing.T,
	use string,
	handler func(
		*cobra.Command,
		*configmanager.ConfigManager,
		[]string,
	) error,
) {
	t.Helper()
	// Create and switch to temp directory for test isolation
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	cmd, manager, _ := NewCommandAndManager(t, use)
	manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))

	err := handler(cmd, manager, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, helpers.ErrConfigurationValidationFailed) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
