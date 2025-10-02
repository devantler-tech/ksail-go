package cmd_test

import (
	"testing"
)

// TestStatusCmdIntegration tests the status command with stub mode.
//
//nolint:paralleltest // Cannot use parallel with t.Chdir()
func TestStatusCmdIntegration(t *testing.T) {
	testClusterCommand(t, "status", "status", "Cluster status")
}
