package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListCmdIntegration tests the list command with stub mode.
func TestListCmdIntegration(t *testing.T) {
	t.Parallel()

	distributions := []string{"Kind", "K3d"}

	for _, dist := range distributions {
		t.Run("list_with_"+dist, func(t *testing.T) {
			t.Parallel()

			// Test list command directly with stub mode
			rootCmd := cmd.NewRootCmd("test", "test", "test")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"--stub", "cluster", "list"})
			
			err := rootCmd.Execute()
			require.NoError(t, err, "list should succeed for distribution %s", dist)

			// Verify output contains expected message
			output := out.String()
			assert.Contains(t, output, "Listing")
		})
	}
}
