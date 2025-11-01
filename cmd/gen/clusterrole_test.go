package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenClusterRole tests generating a clusterrole manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenClusterRole(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewClusterRoleCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-clusterrole", "--verb=get,list", "--resource=nodes"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen clusterrole to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
