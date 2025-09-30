package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClusterUpCommand_Success tests successful cluster up command execution flows.
func TestClusterUpCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedOutput []string
		validateFunc   func(t *testing.T, output string)
	}{
		{
			name: "up_with_kind_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-test-cluster")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster created and started successfully",
				"Distribution: Kind",
				"Context: kind-kind-test-cluster",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "up_with_k3d_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionK3d, "k3d-test-cluster")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster created and started successfully",
				"Distribution: K3d",
				"Context: k3d-k3d-test-cluster",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "up_with_eks_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionEKS, "eks-test-cluster")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster created and started successfully",
				"Distribution: EKS",
				"Context: eks-test-cluster",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "up_with_custom_timeout",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-timeout-test")
			},
			args: []string{"--timeout", "10m"},
			expectedOutput: []string{
				"Cluster created and started successfully",
				"Distribution: Kind",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "up_with_override_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "override-test")
			},
			args: []string{"--distribution", "K3d"},
			expectedOutput: []string{
				"Cluster created and started successfully",
				"Distribution: K3d", // Should use CLI override
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "up_with_override_context",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "context-test")
			},
			args: []string{"--context", "custom-context"},
			expectedOutput: []string{
				"Cluster created and started successfully",
				"Context: custom-context", // Should use CLI override
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory for test
			tempDir := t.TempDir()

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Setup test environment
			testCase.setupFunc(t, tempDir)

			// Create and execute cluster up command
			upCmd := cluster.NewUpCmd()
			var output bytes.Buffer
			upCmd.SetOut(&output)
			upCmd.SetErr(&output)
			upCmd.SetArgs(testCase.args)

			// Execute the command
			err = upCmd.Execute()
			require.NoError(
				t,
				err,
				"cluster up command should succeed, output: %s",
				output.String(),
			)

			// Verify expected output strings
			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}

			// Run custom validation
			if testCase.validateFunc != nil {
				testCase.validateFunc(t, outputStr)
			}
		})
	}
}

// TestClusterUpCommand_ErrorCases tests error scenarios for cluster up command.
func TestClusterUpCommand_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedErrMsg string
	}{
		{
			name: "missing_ksail_config",
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup - missing ksail.yaml
			},
			args:           []string{},
			expectedErrMsg: "failed to load cluster configuration",
		},
		{
			name: "invalid_ksail_config",
			setupFunc: func(t *testing.T, tempDir string) {
				// Create invalid ksail.yaml
				err := os.WriteFile(
					filepath.Join(tempDir, "ksail.yaml"),
					[]byte("invalid: yaml: content: ["),
					0o600,
				)
				require.NoError(t, err)
			},
			args:           []string{},
			expectedErrMsg: "failed to load cluster configuration",
		},
		{
			name: "invalid_timeout_format",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "timeout-test")
			},
			args:           []string{"--timeout", "invalid-timeout"},
			expectedErrMsg: "invalid duration",
		},
		{
			name: "invalid_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "dist-test")
			},
			args:           []string{"--distribution", "InvalidDistribution"},
			expectedErrMsg: "invalid distribution",
		},
		{
			name: "validation_error_missing_required_field",
			setupFunc: func(t *testing.T, tempDir string) {
				// Create config with missing required fields
				configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: ""  # Empty name should cause validation error
spec:
  distribution: Kind`
				err := os.WriteFile(
					filepath.Join(tempDir, "ksail.yaml"),
					[]byte(configContent),
					0o600,
				)
				require.NoError(t, err)
			},
			args:           []string{},
			expectedErrMsg: "Configuration validation failed",
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory for test
			tempDir := t.TempDir()

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Setup test environment
			testCase.setupFunc(t, tempDir)

			// Create and execute cluster up command
			upCmd := cluster.NewUpCmd()
			var output bytes.Buffer
			upCmd.SetOut(&output)
			upCmd.SetErr(&output)
			upCmd.SetArgs(testCase.args)

			// Execute the command and expect an error
			err = upCmd.Execute()
			require.Error(t, err, "cluster up command should fail")

			// Verify error message contains expected text
			if testCase.expectedErrMsg != "" {
				assert.Contains(t, err.Error(), testCase.expectedErrMsg,
					"Error message should contain expected text, output: %s", output.String())
			}
		})
	}
}

// TestClusterUpCommand_HelpAndValidation tests help output and flag validation.
func TestClusterUpCommand_HelpAndValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "help_flag",
			args:           []string{"--help"},
			expectedOutput: "Start the Kubernetes cluster",
			expectError:    false,
		},
		{
			name:           "help_flag_short",
			args:           []string{"-h"},
			expectedOutput: "Start the Kubernetes cluster",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create and execute cluster up command
			upCmd := cluster.NewUpCmd()
			var output bytes.Buffer
			upCmd.SetOut(&output)
			upCmd.SetErr(&output)
			upCmd.SetArgs(testCase.args)

			// Execute the command
			err := upCmd.Execute()
			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err, "command should not return error")
			}

			// Verify expected output
			assert.Contains(t, output.String(), testCase.expectedOutput)
		})
	}
}

// createKSailConfig creates a valid ksail.yaml configuration file for testing.
func createKSailConfig(
	t *testing.T,
	tempDir string,
	distribution v1alpha1.Distribution,
	clusterName string,
) {
	t.Helper()

	var context string
	switch distribution {
	case v1alpha1.DistributionKind:
		context = "kind-" + clusterName
	case v1alpha1.DistributionK3d:
		context = "k3d-" + clusterName
	case v1alpha1.DistributionEKS:
		context = clusterName // EKS doesn't use prefix
	}

	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: ` + clusterName + `
spec:
  distribution: ` + string(distribution) + `
  connection:
    context: ` + context + `
    timeout: 5m
  sourceDirectory: k8s`

	err := os.WriteFile(filepath.Join(tempDir, "ksail.yaml"), []byte(configContent), 0o600)
	require.NoError(t, err)
}
