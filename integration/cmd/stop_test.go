package cmd_test

import (
	"testing"
)

// TestStopCmdIntegration tests the stop command with stub mode.
//
//nolint:paralleltest // Cannot use parallel with t.Chdir()
func TestStopCmdIntegration(t *testing.T) {
	testClusterCommand(t, "stop", "stop", "Cluster stopped successfully")
}
