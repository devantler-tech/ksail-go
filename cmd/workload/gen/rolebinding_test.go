package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenRoleBinding tests generating a rolebinding manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenRoleBinding(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewRoleBindingCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-rolebinding", "--role=test-role", "--user=test-user"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen rolebinding to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
