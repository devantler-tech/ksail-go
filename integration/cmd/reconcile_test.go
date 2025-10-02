package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/require"
)

// TestReconcileCmdIntegration tests the reconcile command with stub mode.
func TestReconcileCmdIntegration(t *testing.T) {
	t.Parallel()

	distributions := []string{"Kind", "K3d"}

	for _, dist := range distributions {
		t.Run("reconcile_with_"+dist, func(t *testing.T) {
			t.Parallel()

			// Test reconcile command directly with stub mode
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "workload", "reconcile"})

			err := rootCmd.Execute()
			require.NoError(t, err, "reconcile should succeed for distribution %s", dist)

			// Reconcile command just prints info message, no specific assertion needed
		})
	}
}
