// Package cmd provides tests for command functionality.
package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestSystemWorkflow_AllCommandsExecuteSuccessfully(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	commands := [][]string{
		{"init", "--container-engine", "Docker", "--distribution", "Kind", "--deployment-tool", "Kubectl"},
		{"up"},
		{"status"},
		{"list"},
		{"list", "--all"},
		{"stop"},
		{"start"},
		{"update"},
		{"down"},
		{"down"}, // Second down should also succeed
	}

	// Act & Assert - Test that all commands in the system workflow execute without error
	for _, cmdArgs := range commands {
		out.Reset()

		rootCmd := cmd.NewRootCmd("test", "test", "test")
		rootCmd.SetOut(&out)
		rootCmd.SetArgs(cmdArgs)

		err := rootCmd.Execute()
		if err != nil {
			test.Fatalf("Command '%v' failed with error: %v", cmdArgs, err)
		}

		output := out.String()
		if output == "" {
			test.Fatalf("Command '%v' produced no output", cmdArgs)
		}

		// Verify that stub implementation messages are present
		if !containsStubMessage(output) {
			test.Fatalf("Command '%v' output does not contain expected stub message: %s", cmdArgs, output)
		}
	}
}

func TestInitCommand_AcceptsAllMatrixFlags(test *testing.T) {
	// Arrange
	test.Parallel()

	testCases := getInitTestCases()

	// Act & Assert
	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()

			var out bytes.Buffer

			rootCmd := cmd.NewRootCmd("test", "test", "test")
			rootCmd.SetOut(&out)
			rootCmd.SetArgs(testCase.args)

			err := rootCmd.Execute()
			if err != nil {
				test.Fatalf("Init command with args '%v' failed: %v", testCase.args, err)
			}

			output := out.String()
			if !containsStubMessage(output) {
				test.Fatalf("Init command output does not contain expected stub message: %s", output)
			}
		})
	}
}

func getInitTestCases() []struct {
	name string
	args []string
} {
	return []struct {
		name string
		args []string
	}{
		{
			name: "Docker Kind Kubectl",
			args: []string{"init", "--container-engine", "Docker", "--distribution", "Kind", "--deployment-tool", "Kubectl"},
		},
		{
			name: "Docker Kind Kubectl with SOPS",
			args: []string{
				"init", "--container-engine", "Docker", "--distribution", "Kind",
				"--deployment-tool", "Kubectl", "--secret-manager", "SOPS",
			},
		},
		{
			name: "Podman K3d",
			args: []string{"init", "--container-engine", "Podman", "--distribution", "K3d"},
		},
		{
			name: "Kind with Cilium CNI",
			args: []string{"init", "--container-engine", "Docker", "--distribution", "Kind", "--cni", "Cilium"},
		},
		{
			name: "Kind with LocalPathProvisioner CSI",
			args: []string{"init", "--container-engine", "Docker", "--distribution", "Kind", "--csi", "LocalPathProvisioner"},
		},
		{
			name: "Kind with Traefik Ingress",
			args: []string{"init", "--container-engine", "Docker", "--distribution", "Kind", "--ingress-controller", "Traefik"},
		},
		{
			name: "Kind with metrics server enabled",
			args: []string{"init", "--container-engine", "Docker", "--distribution", "Kind", "--metrics-server", "True"},
		},
		{
			name: "Kind with mirror registries enabled",
			args: []string{"init", "--container-engine", "Docker", "--distribution", "Kind", "--mirror-registries", "True"},
		},
	}
}

// containsStubMessage checks if the output contains a stub implementation message.
func containsStubMessage(output string) bool {
	stubMessages := []string{
		"stub implementation",
		"Project initialized successfully",
		"Cluster started successfully",
		"Cluster status: Running",
		"Listing",
		"Cluster stopped successfully",
		"Cluster updated successfully",
		"Cluster stopped and removed successfully",
	}

	for _, msg := range stubMessages {
		if strings.Contains(output, msg) {
			return true
		}
	}

	return false
}