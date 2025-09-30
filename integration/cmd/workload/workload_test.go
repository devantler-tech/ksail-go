package integration

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/workload"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkloadApplyCommand_Success tests successful workload apply command execution flows.
func TestWorkloadApplyCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name:           "apply_basic",
			args:           []string{},
			expectedOutput: []string{"Workload apply coming soon."},
		},
		{
			name:           "apply_with_args",
			args:           []string{"--some-flag", "value"},
			expectedOutput: []string{"Workload apply coming soon."},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			
			// Create and execute workload apply command
			applyCmd := workload.NewApplyCommand()
			var output bytes.Buffer
			applyCmd.SetOut(&output)
			applyCmd.SetErr(&output)
			applyCmd.SetArgs(testCase.args)
			
			// Execute the command
			err := applyCmd.Execute()
			require.NoError(t, err, "workload apply command should succeed, output: %s", output.String())
			
			// Verify expected output strings
			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestWorkloadInstallCommand_Success tests successful workload install command execution flows.
func TestWorkloadInstallCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name:           "install_basic",
			args:           []string{},
			expectedOutput: []string{"Workload install coming soon."},
		},
		{
			name:           "install_with_args",
			args:           []string{"--some-flag", "value"},
			expectedOutput: []string{"Workload install coming soon."},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			
			// Create and execute workload install command
			installCmd := workload.NewInstallCommand()
			var output bytes.Buffer
			installCmd.SetOut(&output)
			installCmd.SetErr(&output)
			installCmd.SetArgs(testCase.args)
			
			// Execute the command
			err := installCmd.Execute()
			require.NoError(t, err, "workload install command should succeed, output: %s", output.String())
			
			// Verify expected output strings
			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestWorkloadReconcileCommand_Success tests successful workload reconcile command execution flows.
func TestWorkloadReconcileCommand_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name:           "reconcile_basic",
			args:           []string{},
			expectedOutput: []string{"Workload reconciliation coming soon."},
		},
		{
			name:           "reconcile_with_args",
			args:           []string{"--some-flag", "value"},
			expectedOutput: []string{"Workload reconciliation coming soon."},
		},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			
			// Create and execute workload reconcile command
			reconcileCmd := workload.NewReconcileCommand()
			var output bytes.Buffer
			reconcileCmd.SetOut(&output)
			reconcileCmd.SetErr(&output)
			reconcileCmd.SetArgs(testCase.args)
			
			// Execute the command
			err := reconcileCmd.Execute()
			require.NoError(t, err, "workload reconcile command should succeed, output: %s", output.String())
			
			// Verify expected output strings
			outputStr := output.String()
			for _, expectedStr := range testCase.expectedOutput {
				assert.Contains(t, outputStr, expectedStr, "Output should contain expected string")
			}
		})
	}
}

// TestWorkloadCommands_Help tests help output for all workload commands.
func TestWorkloadCommands_Help(t *testing.T) {
	t.Parallel()

	commands := []struct {
		name           string
		cmdFunc        func() *cobra.Command
		expectedOutput string
	}{
		{
			name:           "apply_help",
			cmdFunc:        workload.NewApplyCommand,
			expectedOutput: "Apply manifests",
		},
		{
			name:           "install_help",
			cmdFunc:        workload.NewInstallCommand,
			expectedOutput: "Install Helm charts",
		},
		{
			name:           "reconcile_help",
			cmdFunc:        workload.NewReconcileCommand,
			expectedOutput: "Reconcile workloads with the cluster",
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

		t.Run(cmdTest.name+"_short", func(t *testing.T) {
			t.Parallel()
			
			cmd := cmdTest.cmdFunc()
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{"-h"})
			
			err := cmd.Execute()
			require.NoError(t, err, "help command should not return error")
			assert.Contains(t, output.String(), cmdTest.expectedOutput)
		})
	}
}

// TestWorkloadCommands_Integration tests end-to-end integration scenarios.
func TestWorkloadCommands_Integration(t *testing.T) {
	t.Parallel()

	// Test that all workload commands can be executed without errors
	// and produce expected output regardless of environment setup
	tests := []struct {
		name           string
		cmdFunc        func() *cobra.Command
		expectedOutput string
	}{
		{
			name:           "apply_integration",
			cmdFunc:        workload.NewApplyCommand,
			expectedOutput: "Workload apply coming soon.",
		},
		{
			name:           "install_integration",
			cmdFunc:        workload.NewInstallCommand,
			expectedOutput: "Workload install coming soon.",
		},
		{
			name:           "reconcile_integration",
			cmdFunc:        workload.NewReconcileCommand,
			expectedOutput: "Workload reconciliation coming soon.",
		},
	}

	for _, cmdTest := range tests {
		cmdTest := cmdTest
		t.Run(cmdTest.name, func(t *testing.T) {
			t.Parallel()
			
			cmd := cmdTest.cmdFunc()
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{})
			
			err := cmd.Execute()
			require.NoError(t, err, "command should succeed")
			assert.Contains(t, output.String(), cmdTest.expectedOutput)
		})
	}
}

// TestWorkloadCommands_MultipleExecutions tests that commands can be executed multiple times.
func TestWorkloadCommands_MultipleExecutions(t *testing.T) {
	t.Parallel()

	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{
			name:    "apply_multiple",
			cmdFunc: workload.NewApplyCommand,
		},
		{
			name:    "install_multiple",
			cmdFunc: workload.NewInstallCommand,
		},
		{
			name:    "reconcile_multiple",
			cmdFunc: workload.NewReconcileCommand,
		},
	}

	for _, cmdTest := range commands {
		cmdTest := cmdTest
		t.Run(cmdTest.name, func(t *testing.T) {
			t.Parallel()
			
			// Execute the same command multiple times to ensure no state issues
			for i := 0; i < 3; i++ {
				cmd := cmdTest.cmdFunc()
				var output bytes.Buffer
				cmd.SetOut(&output)
				cmd.SetErr(&output)
				cmd.SetArgs([]string{})
				
				err := cmd.Execute()
				require.NoError(t, err, "command execution %d should succeed", i+1)
				assert.NotEmpty(t, output.String(), "output should not be empty on execution %d", i+1)
			}
		})
	}
}