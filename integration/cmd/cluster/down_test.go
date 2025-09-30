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

// TestClusterDownCommand_Success tests successful cluster down command execution flows.
func TestClusterDownCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedOutput []string
		validateFunc   func(t *testing.T, output string)
	}{
		{
			name: "down_with_kind_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-down-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster destroyed successfully",
				"Distribution: Kind",
				"Context: kind-kind-down-test",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "down_with_k3d_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionK3d, "k3d-down-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster destroyed successfully",
				"Distribution: K3d",
				"Context: k3d-k3d-down-test",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "down_with_eks_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionEKS, "eks-down-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster destroyed successfully",
				"Distribution: EKS",
				"Context: eks-down-test",
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "down_with_override_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "override-down-test")
			},
			args: []string{"--distribution", "K3d"},
			expectedOutput: []string{
				"Cluster destroyed successfully",
				"Distribution: K3d", // Should use CLI override
			},
			validateFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "✔")
			},
		},
		{
			name: "down_with_override_context",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "context-down-test")
			},
			args: []string{"--context", "custom-down-context"},
			expectedOutput: []string{
				"Cluster destroyed successfully",
				"Context: custom-down-context", // Should use CLI override
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

			// Create and execute cluster down command
			downCmd := cluster.NewDownCmd()
			var output bytes.Buffer
			downCmd.SetOut(&output)
			downCmd.SetErr(&output)
			downCmd.SetArgs(testCase.args)

			// Execute the command
			err = downCmd.Execute()
			require.NoError(
				t,
				err,
				"cluster down command should succeed, output: %s",
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

// TestClusterDownCommand_ErrorCases tests error scenarios for cluster down command.
func TestClusterDownCommand_ErrorCases(t *testing.T) {
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
			name: "invalid_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "dist-down-test")
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
		{
			name: "empty_context_validation",
			setupFunc: func(t *testing.T, tempDir string) {
				// Create config with empty context
				configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: context-test
spec:
  distribution: Kind
  connection:
    context: ""  # Empty context might cause validation error
    timeout: 5m`
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

			// Create and execute cluster down command
			downCmd := cluster.NewDownCmd()
			var output bytes.Buffer
			downCmd.SetOut(&output)
			downCmd.SetErr(&output)
			downCmd.SetArgs(testCase.args)

			// Execute the command and expect an error
			err = downCmd.Execute()
			require.Error(t, err, "cluster down command should fail")

			// Verify error message contains expected text
			if testCase.expectedErrMsg != "" {
				assert.Contains(t, err.Error(), testCase.expectedErrMsg,
					"Error message should contain expected text, output: %s", output.String())
			}
		})
	}
}

// TestClusterDownCommand_HelpAndValidation tests help output for cluster down command.
func TestClusterDownCommand_HelpAndValidation(t *testing.T) {
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
			expectedOutput: "Destroy a cluster",
			expectError:    false,
		},
		{
			name:           "help_flag_short",
			args:           []string{"-h"},
			expectedOutput: "Destroy a cluster",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create and execute cluster down command
			downCmd := cluster.NewDownCmd()
			var output bytes.Buffer
			downCmd.SetOut(&output)
			downCmd.SetErr(&output)
			downCmd.SetArgs(testCase.args)

			// Execute the command
			err := downCmd.Execute()
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
