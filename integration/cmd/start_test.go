package cmd_test

import (
	"testing"
)

// TestStartCmdIntegration tests the start command with stub mode.
//
//nolint:paralleltest // Cannot use parallel with t.Chdir()
func TestStartCmdIntegration(t *testing.T) {
	testClusterCommand(t, "start", "start", "Cluster started successfully")
}
