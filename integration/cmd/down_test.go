package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDownCmdIntegration tests the down command with stub mode.
//
//nolint:paralleltest // Cannot use parallel with t.Chdir()
func TestDownCmdIntegration(t *testing.T) {
	tests := []struct {
		distribution string
		config       string
	}{
		{distribution: "Kind", config: "kind.yaml"},
		{distribution: "K3d", config: "k3d.yaml"},
	}

	for _, testCase := range tests {
		t.Run("down_with_"+testCase.distribution, func(t *testing.T) {
			// Commands now work with defaults from field selectors - no config file needed
			tempDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			t.Chdir(tempDir)

			defer func() {
				//nolint:usetesting // Cleanup requires restoring original directory
				_ = os.Chdir(origDir)
			}()

			// Test down command with stub mode
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{
				"--stub",
				"cluster",
				"down",
				"--distribution", testCase.distribution,
				"--distribution-config", testCase.config,
			})

			err = rootCmd.Execute()
			require.NoError(
				t,
				err,
				"down should succeed for distribution %s",
				testCase.distribution,
			)

			// Verify output contains expected message
			output := out.String()
			assert.Contains(t, output, "cluster destroyed successfully")
		})
	}
}
