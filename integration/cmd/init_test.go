package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitCommand_Success tests successful init command execution flows.
func TestInitCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		args            []string
		distribution    v1alpha1.Distribution
		expectedFiles   []string
		validateContent func(t *testing.T, tempDir string)
	}{
		{
			name:         "init_with_kind_distribution",
			args:         []string{"--distribution", "Kind"},
			distribution: v1alpha1.DistributionKind,
			expectedFiles: []string{
				"ksail.yaml",
				"kind.yaml",
				"k8s/kustomization.yaml",
			},
			validateContent: func(t *testing.T, tempDir string) {
				// Validate ksail.yaml contains Kind distribution
				ksailContent, err := os.ReadFile(filepath.Join(tempDir, "ksail.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(ksailContent), "distribution: Kind")

				// Validate kind.yaml exists and has content
				kindContent, err := os.ReadFile(filepath.Join(tempDir, "kind.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(kindContent), "kind: Cluster")
			},
		},
		{
			name:         "init_with_k3d_distribution",
			args:         []string{"--distribution", "K3d"},
			distribution: v1alpha1.DistributionK3d,
			expectedFiles: []string{
				"ksail.yaml",
				"k3d.yaml",
				"k8s/kustomization.yaml",
			},
			validateContent: func(t *testing.T, tempDir string) {
				// Validate ksail.yaml contains K3d distribution
				ksailContent, err := os.ReadFile(filepath.Join(tempDir, "ksail.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(ksailContent), "distribution: K3d")

				// Validate k3d.yaml exists and has content
				k3dContent, err := os.ReadFile(filepath.Join(tempDir, "k3d.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(k3dContent), "apiVersion:")
			},
		},
		{
			name:         "init_with_eks_distribution",
			args:         []string{"--distribution", "EKS"},
			distribution: v1alpha1.DistributionEKS,
			expectedFiles: []string{
				"ksail.yaml",
				"eks.yaml",
				"k8s/kustomization.yaml",
			},
			validateContent: func(t *testing.T, tempDir string) {
				// Validate ksail.yaml contains EKS distribution
				ksailContent, err := os.ReadFile(filepath.Join(tempDir, "ksail.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(ksailContent), "distribution: EKS")

				// Validate eks.yaml exists and has content
				eksContent, err := os.ReadFile(filepath.Join(tempDir, "eks.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(eksContent), "apiVersion: eksctl.io")
			},
		},
		{
			name:         "init_with_custom_output_directory",
			args:         []string{"--distribution", "Kind", "--output", "custom-project"},
			distribution: v1alpha1.DistributionKind,
			expectedFiles: []string{
				"custom-project/ksail.yaml",
				"custom-project/kind.yaml",
				"custom-project/k8s/kustomization.yaml",
			},
			validateContent: func(t *testing.T, tempDir string) {
				// Validate files are created in custom directory
				ksailContent, err := os.ReadFile(
					filepath.Join(tempDir, "custom-project", "ksail.yaml"),
				)
				require.NoError(t, err)
				assert.Contains(t, string(ksailContent), "distribution: Kind")
			},
		},
		{
			name:         "init_with_force_flag",
			args:         []string{"--distribution", "Kind", "--force"},
			distribution: v1alpha1.DistributionKind,
			expectedFiles: []string{
				"ksail.yaml",
				"kind.yaml",
				"k8s/kustomization.yaml",
			},
			validateContent: func(t *testing.T, tempDir string) {
				// Validate force flag allows overwriting existing files
				ksailContent, err := os.ReadFile(filepath.Join(tempDir, "ksail.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(ksailContent), "distribution: Kind")
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

			// For force flag test, create existing files
			if contains(testCase.args, "--force") {
				err := os.WriteFile("ksail.yaml", []byte("existing content"), 0o600)
				require.NoError(t, err)
			}

			// Create and execute init command
			initCmd := cmd.NewInitCmd()
			var output bytes.Buffer
			initCmd.SetOut(&output)
			initCmd.SetErr(&output)
			initCmd.SetArgs(testCase.args)

			// Execute the command
			err = initCmd.Execute()
			require.NoError(t, err, "init command should succeed, output: %s", output.String())

			// Verify expected files were created
			for _, expectedFile := range testCase.expectedFiles {
				filePath := filepath.Join(tempDir, expectedFile)
				assert.FileExists(t, filePath, "Expected file should exist: %s", expectedFile)
			}

			// Run custom validation
			if testCase.validateContent != nil {
				testCase.validateContent(t, tempDir)
			}

			// Verify success message in output
			assert.Contains(t, output.String(), "Project initialized successfully")
		})
	}
}

// TestInitCommand_ErrorCases tests error scenarios for the init command.
func TestInitCommand_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		setupFunc      func(t *testing.T, tempDir string)
		expectedErrMsg string
	}{
		{
			name:           "invalid_distribution",
			args:           []string{"--distribution", "InvalidDistribution"},
			expectedErrMsg: "invalid distribution",
		},
		{
			name: "existing_files_without_force",
			args: []string{"--distribution", "Kind"},
			setupFunc: func(t *testing.T, tempDir string) {
				// Create existing ksail.yaml
				err := os.WriteFile(filepath.Join(tempDir, "ksail.yaml"), []byte("existing"), 0o600)
				require.NoError(t, err)
			},
			expectedErrMsg: "file already exists",
		},
		{
			name:           "invalid_output_directory_path",
			args:           []string{"--distribution", "Kind", "--output", "/invalid/\x00/path"},
			expectedErrMsg: "invalid path",
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

			// Run setup function if provided
			if testCase.setupFunc != nil {
				testCase.setupFunc(t, tempDir)
			}

			// Create and execute init command
			initCmd := cmd.NewInitCmd()
			var output bytes.Buffer
			initCmd.SetOut(&output)
			initCmd.SetErr(&output)
			initCmd.SetArgs(testCase.args)

			// Execute the command and expect an error
			err = initCmd.Execute()
			require.Error(t, err, "init command should fail")

			// Verify error message contains expected text
			if testCase.expectedErrMsg != "" {
				assert.Contains(t, err.Error(), testCase.expectedErrMsg,
					"Error message should contain expected text, output: %s", output.String())
			}
		})
	}
}

// TestInitCommand_HelpAndVersion tests help and version output for init command.
func TestInitCommand_HelpAndVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "help_flag",
			args:           []string{"--help"},
			expectedOutput: "Initialize a new project",
		},
		{
			name:           "help_flag_short",
			args:           []string{"-h"},
			expectedOutput: "Initialize a new project",
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create and execute init command
			initCmd := cmd.NewInitCmd()
			var output bytes.Buffer
			initCmd.SetOut(&output)
			initCmd.SetErr(&output)
			initCmd.SetArgs(testCase.args)

			// Execute the command (help should not return error)
			err := initCmd.Execute()
			require.NoError(t, err, "help command should not return error")

			// Verify expected output
			assert.Contains(t, output.String(), testCase.expectedOutput)
		})
	}
}

// Helper function to check if slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
