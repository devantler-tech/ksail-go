package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClusterCommand is a helper function to test cluster commands with stub mode.
// It reduces duplication across status, stop, start, and down test files.
func testClusterCommand(t *testing.T, testName, commandName, expectedOutput string) {
	t.Helper()

	tests := []struct {
		distribution string
		config       string
		context      string
	}{
		{distribution: "Kind", config: "kind.yaml", context: "kind-kind"},
		{distribution: "K3d", config: "k3d.yaml", context: "k3d-k3s-default"},
	}

	for _, testCase := range tests {
		t.Run(testName+"_with_"+testCase.distribution, func(t *testing.T) {
			// Commands now work with defaults from field selectors - no config file needed
			tempDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			t.Chdir(tempDir)

			defer func() {
				//nolint:usetesting // Cleanup requires restoring original directory
				_ = os.Chdir(origDir)
			}()

			// Test cluster command with stub mode
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{
				"--stub",
				"cluster",
				commandName,
				"--distribution", testCase.distribution,
				"--distribution-config", testCase.config,
				"--context", testCase.context,
			})

			err = rootCmd.Execute()
			require.NoError(
				t,
				err,
				"%s should succeed for distribution %s",
				commandName,
				testCase.distribution,
			)

			// Verify output contains expected message
			output := out.String()
			assert.Contains(t, output, expectedOutput)
		})
	}
}
