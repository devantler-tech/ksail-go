package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClusterLifecycleIntegration tests the full cluster lifecycle with stub mode.
// Note: Not running in parallel due to potential global state interactions in config loading.
//
//nolint:funlen,paralleltest // Integration test intentionally covers full lifecycle and avoids parallelism
func TestClusterLifecycleIntegration(t *testing.T) {
	distributions := []string{"Kind", "K3d"}

	for _, dist := range distributions { //nolint:paralleltest // Intentionally sequential to avoid race conditions
		t.Run("lifecycle_"+dist, func(t *testing.T) {
			// Don't run subtests in parallel either to avoid race conditions
			// t.Parallel()

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

			// Step 1: Initialize project
			t.Log("Step 1: Init")

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
			assert.Contains(t, out.String(), "initialized project")

			// Step 2: Up (create cluster)
			t.Log("Step 2: Up")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "up"})
			err = rootCmd.Execute()
			require.NoError(t, err, "up should succeed")
			assert.Contains(t, out.String(), "Cluster created and started successfully")

			// Step 3: Status
			t.Log("Step 3: Status")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "status"})
			err = rootCmd.Execute()
			require.NoError(t, err, "status should succeed")
			assert.Contains(t, out.String(), "Cluster status")

			// Step 4: List
			t.Log("Step 4: List")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "list"})
			err = rootCmd.Execute()
			require.NoError(t, err, "list should succeed")
			assert.Contains(t, out.String(), "Listing")

			// Step 5: Stop
			t.Log("Step 5: Stop")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "stop"})
			err = rootCmd.Execute()
			require.NoError(t, err, "stop should succeed")
			assert.Contains(t, out.String(), "Cluster stopped successfully")

			// Step 6: Start
			t.Log("Step 6: Start")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "start"})
			err = rootCmd.Execute()
			require.NoError(t, err, "start should succeed")
			assert.Contains(t, out.String(), "Cluster started successfully")

			// Step 7: Reconcile workload
			t.Log("Step 7: Reconcile")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "workload", "reconcile"})
			err = rootCmd.Execute()
			require.NoError(t, err, "reconcile should succeed")
			// Reconcile just prints info message

			// Step 8: Down (destroy cluster)
			t.Log("Step 8: Down")

			rootCmd = cmd.NewRootCmd("test", "test", "test")

			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "down"})
			err = rootCmd.Execute()
			require.NoError(t, err, "down should succeed")
			assert.Contains(t, out.String(), "cluster destroyed successfully")

			// Verify files exist (check in actual temp directory, not current working directory)
			assert.FileExists(t, filepath.Join(tempDir, "ksail.yaml"))
			// Distribution config files are lowercase (kind.yaml, k3d.yaml)
			distFile := filepath.Join(tempDir, strings.ToLower(dist)+".yaml")
			assert.FileExists(t, distFile)
		})
	}
}
