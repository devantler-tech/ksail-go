package cmd_test

import (
	"testing"
)

// TestDownCmdIntegration tests the down command with stub mode.
//
//nolint:paralleltest // Cannot use parallel with t.Chdir()
func TestDownCmdIntegration(t *testing.T) {
	testClusterCommand(t, "down", "down", "cluster destroyed successfully")
}
