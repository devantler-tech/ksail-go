package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCmdIntegration tests the status command with stub mode.
func TestStatusCmdIntegration(t *testing.T) {
	distributions := []string{"Kind", "K3d"}

	for _, dist := range distributions {
		t.Run("status_with_"+dist, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			// Change to temp directory
			origDir, err := os.Getwd()
			require.NoError(t, err)

			t.Chdir(tempDir)

			defer func() {
				//nolint:usetesting // Cleanup requires restoring original directory
				_ = os.Chdir(origDir)
			}()

			// First initialize project
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{
				"--stub",
				"init",
				"--distribution", dist,
			})
			err = rootCmd.Execute()
			require.NoError(t, err, "init should succeed")

			// Test status command
			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "status"})
			err = rootCmd.Execute()
			require.NoError(t, err, "status should succeed for distribution %s", dist)

			// Verify output
			output := out.String()
			assert.Contains(t, output, "Cluster status")
		})
	}
}
