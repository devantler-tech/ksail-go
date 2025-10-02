package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStopCmdIntegration tests the stop command with stub mode.
func TestStopCmdIntegration(t *testing.T) {
	tests := []struct {
		distribution string
		config       string
	}{
		{distribution: "Kind", config: "kind.yaml"},
		{distribution: "K3d", config: "k3d.yaml"},
	}

	for _, tt := range tests {
		t.Run("stop_with_"+tt.distribution, func(t *testing.T) {
			// Commands now work with defaults from field selectors - no config file needed
			tempDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			t.Chdir(tempDir)

			defer func() {
				//nolint:usetesting // Cleanup requires restoring original directory
				_ = os.Chdir(origDir)
			}()

			// Test stop command with stub mode
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{
				"--stub",
				"cluster",
				"stop",
				"--distribution", tt.distribution,
				"--distribution-config", tt.config,
			})

			err = rootCmd.Execute()
			require.NoError(t, err, "stop should succeed for distribution %s", tt.distribution)

			// Verify output contains expected message
			output := out.String()
			assert.Contains(t, output, "Cluster stopped successfully")
		})
	}
}
