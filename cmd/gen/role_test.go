package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenRole tests generating a role manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenRole(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewRoleCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-role", "--verb=get,list", "--resource=pods"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen role to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
