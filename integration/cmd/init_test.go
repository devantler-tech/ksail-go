package cmd_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitCmdIntegration tests the init command with stub mode.
func TestInitCmdIntegration(t *testing.T) {
	t.Parallel()

	distributions := []string{"Kind", "K3d"}

	for _, dist := range distributions {
		t.Run("init_with_"+dist, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory for test
			tempDir := t.TempDir()

			// Create init command
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)

			// Set arguments for init command with stub mode
			rootCmd.SetArgs([]string{
				"--stub",
				"init",
				"--distribution", dist,
				"--output", tempDir,
			})

			// Execute command
			err := rootCmd.Execute()
			require.NoError(t, err, "init command should succeed for distribution %s", dist)

			// Verify output contains expected messages
			output := out.String()
			assert.Contains(t, output, "Initializing project",
				"output should contain initialization message")
			assert.Contains(t, output, "initialized project",
				"output should contain success message")
			assert.Contains(t, output, "STUB:", "output should indicate stub mode")

			// Verify stub files were created
			assert.FileExists(t, filepath.Join(tempDir, "ksail.yaml"),
				"ksail.yaml should be created")
			// Distribution config files are lowercase (kind.yaml, k3d.yaml)
			distFile := util.DistributionFileName(dist)
			assert.FileExists(
				t,
				filepath.Join(tempDir, distFile),
				"distribution config should be created",
			)
			assert.DirExists(t, filepath.Join(tempDir, "k8s"), "k8s directory should be created")
			assert.FileExists(
				t,
				filepath.Join(tempDir, "k8s", "kustomization.yaml"),
				"kustomization.yaml should be created",
			)
		})
	}
}
