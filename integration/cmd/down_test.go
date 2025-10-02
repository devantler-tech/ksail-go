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
func TestDownCmdIntegration(t *testing.T) {
	distributions := []string{"Kind", "K3d"}

	for _, dist := range distributions {
		t.Run("down_with_"+dist, func(t *testing.T) {
			// Create minimal config file for commands that validate configuration
			tempDir := t.TempDir()
			origDir, err := os.Getwd()
			require.NoError(t, err)
			t.Chdir(tempDir)
			defer func() {
				//nolint:usetesting // Cleanup requires restoring original directory
				_ = os.Chdir(origDir)
			}()

			// Determine correct context for distribution
			var context string
			if dist == "Kind" {
				context = "kind-kind"
			} else {
				context = "k3d-k3s-default"
			}

			// Create minimal valid ksail.yaml
			configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: ` + dist + `
  distributionConfig: ` + dist + `.yaml
  connection:
    context: ` + context + `
  sourceDirectory: k8s
`
			err = os.WriteFile("ksail.yaml", []byte(configContent), 0o600)
			require.NoError(t, err)

			// Test down command with stub mode
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "down"})

			err = rootCmd.Execute()
			require.NoError(t, err, "down should succeed for distribution %s", dist)

			// Verify output contains expected message
			output := out.String()
			assert.Contains(t, output, "cluster destroyed successfully")
		})
	}
}
