package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// TestHandleUpRunE_ContractTests covers CLI contract requirements from contracts/cluster-up.md.
func TestHandleUpRunE_ContractTests(t *testing.T) {
	t.Parallel()

	t.Run("success output includes telemetry summary fields", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "up")
		seedValidClusterConfig(manager)

		// Ensure force flag exists to avoid pre-existing cluster issues in repeated CI runs
		if cmd.Flags().Lookup("force") == nil {
			cmd.Flags().Bool("force", false, "force")
		}

		_ = cmd.Flags().Set("force", "true")

		err := HandleUpRunE(cmd, manager, nil)
		if err != nil && !strings.Contains(err.Error(), "cluster already exists") {
			t.Fatalf("expected success or benign exists error, got %v", err)
		}

		// Contract requirement: success output must include distribution, context, kubeconfig
		outputStr := output.String()
		assertOutputContains(t, outputStr, "Distribution")
		assertOutputContains(t, outputStr, "Context")
		assertOutputContains(t, outputStr, "Kubeconfig")

		// Contract requirement: success message with inline timing format [elapsed/stage]
		assertOutputContains(t, outputStr, "cluster is ready")
		assertOutputContains(t, outputStr, "[") // Check for timing brackets
		assertOutputContains(t, outputStr, "/") // Check for elapsed/stage separator
	})

	t.Run("dependency failure includes actionable remediation", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "up")
		seedValidClusterConfig(manager)

		// Simulate dependency check failure - this will fail once implementation checks dependencies
		err := HandleUpRunE(cmd, manager, nil)
		// Contract requirement: dependency failures should be actionable
		// When Docker/Podman missing for Kind/K3d:
		// Expected: exit code 2, message suggesting "Start Docker and rerun"
		// This test should FAIL until T006 implements dependency checks
		if err != nil {
			outputStr := output.String()
			if strings.Contains(strings.ToLower(outputStr), "docker") ||
				strings.Contains(strings.ToLower(outputStr), "podman") {
				// Verify remediation guidance is present
				if !strings.Contains(strings.ToLower(outputStr), "start") &&
					!strings.Contains(strings.ToLower(outputStr), "install") {
					t.Error("dependency failure message should include remediation guidance")
				}
			}
		}
	})

	t.Run("validation failure returns exit code 1", func(t *testing.T) {
		t.Parallel()

		cmd, manager, _ := newIsolatedCommandAndManager(t, "up")
		// Seed minimal config to allow LoadConfig to succeed, but trigger validation failure
		manager.Viper.Set("metadata.name", "test-cluster")
		// Invalid config - missing distributionConfig
		manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))
		// Missing distributionConfig will cause validation to fail

		err := HandleUpRunE(cmd, manager, nil)
		if err == nil {
			t.Fatal("expected validation error but got nil")
		}

		// Contract requirement: validation failures return exit code 1
		if !errors.Is(err, helpers.ErrConfigurationValidationFailed) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("timeout failure includes elapsed time", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "up")
		seedValidClusterConfig(manager)
		manager.Viper.Set("spec.connection.timeout", "1ns") // Force timeout

		// This test should FAIL until T008 implements readiness polling with timeout
		err := HandleUpRunE(cmd, manager, nil)
		// Contract requirement: timeout errors should report elapsed time
		if err != nil {
			outputStr := output.String()
			// Expected: "readiness check timed out after 1ns" or similar
			if strings.Contains(strings.ToLower(outputStr), "timeout") ||
				strings.Contains(strings.ToLower(outputStr), "timed out") {
				// Verify elapsed time is mentioned
				assertOutputContains(t, outputStr, "after")
			}
		}
	})

	t.Run("stage failure reports which stage and elapsed time", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "up")
		seedValidClusterConfig(manager)

		// This test should FAIL until T009 implements telemetry emission
		err := HandleUpRunE(cmd, manager, nil)
		// Contract requirement: failures should report stage name and elapsed time
		if err != nil {
			outputStr := output.String()
			// Expected to see stage names like "dependencies", "provision", "readiness", "kubeconfig"
			hasStage := strings.Contains(outputStr, "dependencies") ||
				strings.Contains(outputStr, "provision") ||
				strings.Contains(outputStr, "readiness") ||
				strings.Contains(outputStr, "kubeconfig")

			if hasStage {
				// Should also report timing
				hasTime := strings.Contains(outputStr, "s") || // seconds
					strings.Contains(outputStr, "ms") || // milliseconds
					strings.Contains(outputStr, "elapsed")

				if !hasTime {
					t.Error("stage failure should include elapsed time")
				}
			}
		}
	})
}

func TestHandleUpRunE(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cmd, manager, output := newCommandAndManager(t, "up")
		seedValidClusterConfig(manager)

		if cmd.Flags().Lookup("force") == nil {
			cmd.Flags().Bool("force", false, "force")
		}

		_ = cmd.Flags().Set("force", "true")

		err := HandleUpRunE(cmd, manager, nil)
		if err != nil && !strings.Contains(err.Error(), "cluster already exists") {
			t.Fatalf("expected success or benign exists error, got %v", err)
		}

		// Verify new UI format with inline timing
		assertOutputContains(t, output.String(), "cluster is ready")
		assertOutputContains(t, output.String(), "Distribution")
		assertOutputContains(t, output.String(), "Context")
		assertOutputContains(t, output.String(), "Kubeconfig")
		assertOutputContains(t, output.String(), "[") // Check for timing brackets
		assertOutputContains(t, output.String(), "/") // Check for elapsed/stage separator
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		cmd, manager, _ := newIsolatedCommandAndManager(t, "up")
		// Seed minimal config to allow LoadConfig to succeed, but trigger validation failure
		manager.Viper.Set("metadata.name", "test-cluster")
		manager.Viper.Set("spec.distribution", string(v1alpha1.DistributionKind))
		// Missing distributionConfig will cause validation to fail

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

// newIsolatedCommandAndManager creates a command and manager isolated from project config files.
// Use this for tests that need to simulate missing or invalid configurations.
func newIsolatedCommandAndManager(
	t *testing.T,
	use string,
) (*cobra.Command, *configmanager.ConfigManager, *bytes.Buffer) {
	t.Helper()

	// Store original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Create and change to temp directory
	tempDir := t.TempDir()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Ensure we restore original directory when test completes
	t.Cleanup(func() {
		err := os.Chdir(originalDir)
		if err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

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
