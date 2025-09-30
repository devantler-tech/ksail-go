package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClusterStartCommand_Success tests successful cluster start command execution flows.
func TestClusterStartCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedOutput []string
	}{
		{
			name: "start_with_kind_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-start-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster started successfully",
				"Distribution: Kind",
				"Context: kind-kind-start-test",
			},
		},
		{
			name: "start_with_k3d_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionK3d, "k3d-start-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster started successfully",
				"Distribution: K3d",
				"Context: k3d-k3d-start-test",
			},
		},
		{
			name: "start_with_eks_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionEKS, "eks-start-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster started successfully",
				"Distribution: EKS",
				"Context: eks-start-test",
			},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			testCase.setupFunc(t, tempDir)

			startCmd := cluster.NewStartCmd()
			var output bytes.Buffer
			startCmd.SetOut(&output)
			startCmd.SetErr(&output)
			startCmd.SetArgs(testCase.args)

			err = startCmd.Execute()
			require.NoError(
				t,
				err,
				"cluster start command should succeed, output: %s",
				output.String(),
			)

			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestClusterStopCommand_Success tests successful cluster stop command execution flows.
func TestClusterStopCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedOutput []string
	}{
		{
			name: "stop_with_kind_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-stop-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster stopped successfully",
				"Distribution: Kind",
				"Context: kind-kind-stop-test",
			},
		},
		{
			name: "stop_with_k3d_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionK3d, "k3d-stop-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster stopped successfully",
				"Distribution: K3d",
				"Context: k3d-k3d-stop-test",
			},
		},
		{
			name: "stop_with_eks_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionEKS, "eks-stop-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster stopped successfully",
				"Distribution: EKS",
				"Context: eks-stop-test",
			},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			testCase.setupFunc(t, tempDir)

			stopCmd := cluster.NewStopCmd()
			var output bytes.Buffer
			stopCmd.SetOut(&output)
			stopCmd.SetErr(&output)
			stopCmd.SetArgs(testCase.args)

			err = stopCmd.Execute()
			require.NoError(
				t,
				err,
				"cluster stop command should succeed, output: %s",
				output.String(),
			)

			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestClusterStatusCommand_Success tests successful cluster status command execution flows.
func TestClusterStatusCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedOutput []string
	}{
		{
			name: "status_with_kind_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-status-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster status: Running",
				"Distribution: Kind",
				"Context: kind-kind-status-test",
			},
		},
		{
			name: "status_with_k3d_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionK3d, "k3d-status-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster status: Running",
				"Distribution: K3d",
				"Context: k3d-k3d-status-test",
			},
		},
		{
			name: "status_with_eks_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionEKS, "eks-status-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Cluster status: Running",
				"Distribution: EKS",
				"Context: eks-status-test",
			},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			testCase.setupFunc(t, tempDir)

			statusCmd := cluster.NewStatusCmd()
			var output bytes.Buffer
			statusCmd.SetOut(&output)
			statusCmd.SetErr(&output)
			statusCmd.SetArgs(testCase.args)

			err = statusCmd.Execute()
			require.NoError(
				t,
				err,
				"cluster status command should succeed, output: %s",
				output.String(),
			)

			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestClusterListCommand_Success tests successful cluster list command execution flows.
func TestClusterListCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		args           []string
		expectedOutput []string
	}{
		{
			name: "list_with_kind_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionKind, "kind-list-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Clusters listed successfully",
				"Distribution: Kind",
				"Context: kind-kind-list-test",
			},
		},
		{
			name: "list_with_k3d_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionK3d, "k3d-list-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Clusters listed successfully",
				"Distribution: K3d",
				"Context: k3d-k3d-list-test",
			},
		},
		{
			name: "list_with_eks_distribution",
			setupFunc: func(t *testing.T, tempDir string) {
				createKSailConfig(t, tempDir, v1alpha1.DistributionEKS, "eks-list-test")
			},
			args: []string{},
			expectedOutput: []string{
				"Clusters listed successfully",
				"Distribution: EKS",
				"Context: eks-list-test",
			},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			testCase.setupFunc(t, tempDir)

			listCmd := cluster.NewListCmd()
			var output bytes.Buffer
			listCmd.SetOut(&output)
			listCmd.SetErr(&output)
			listCmd.SetArgs(testCase.args)

			err = listCmd.Execute()
			require.NoError(
				t,
				err,
				"cluster list command should succeed, output: %s",
				output.String(),
			)

			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestClusterCommands_ErrorCases tests common error scenarios for cluster commands.
func TestClusterCommands_ErrorCases(t *testing.T) {
	t.Parallel()

	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"start", cluster.NewStartCmd},
		{"stop", cluster.NewStopCmd},
		{"status", cluster.NewStatusCmd},
		{"list", cluster.NewListCmd},
	}

	for _, cmdTest := range commands {
		cmdTest := cmdTest
		t.Run(cmdTest.name+"_missing_config", func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			cmd := cmdTest.cmdFunc()
			err = cmd.Execute()
			require.Error(t, err, cmdTest.name+" command should fail when config is missing")
			assert.Contains(t, err.Error(), "failed to load cluster configuration")
		})

		t.Run(cmdTest.name+"_invalid_config", func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(oldWd) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create invalid ksail.yaml
			err = os.WriteFile(
				filepath.Join(tempDir, "ksail.yaml"),
				[]byte("invalid: yaml: content: ["),
				0o600,
			)
			require.NoError(t, err)

			cmd := cmdTest.cmdFunc()
			err = cmd.Execute()
			require.Error(t, err, cmdTest.name+" command should fail with invalid config")
			assert.Contains(t, err.Error(), "failed to load cluster configuration")
		})
	}
}

// TestClusterCommands_Help tests help output for all cluster commands.
func TestClusterCommands_Help(t *testing.T) {
	t.Parallel()

	commands := []struct {
		name           string
		cmdFunc        func() *cobra.Command
		expectedOutput string
	}{
		{
			name:           "start_help",
			cmdFunc:        cluster.NewStartCmd,
			expectedOutput: "Start a previously stopped cluster",
		},
		{
			name:           "stop_help",
			cmdFunc:        cluster.NewStopCmd,
			expectedOutput: "Stop the Kubernetes cluster",
		},
		{
			name:           "status_help",
			cmdFunc:        cluster.NewStatusCmd,
			expectedOutput: "Show the current status",
		},
		{
			name:           "list_help",
			cmdFunc:        cluster.NewListCmd,
			expectedOutput: "List all Kubernetes clusters",
		},
	}

	for _, cmdTest := range commands {
		cmdTest := cmdTest
		t.Run(cmdTest.name, func(t *testing.T) {
			t.Parallel()

			cmd := cmdTest.cmdFunc()
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{"--help"})

			err := cmd.Execute()
			require.NoError(t, err, "help command should not return error")
			assert.Contains(t, output.String(), cmdTest.expectedOutput)
		})
	}
}
