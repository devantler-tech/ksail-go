package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenClusterRoleBinding tests generating a clusterrolebinding manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenClusterRoleBinding(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewClusterRoleBindingCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs(
		[]string{"test-clusterrolebinding", "--clusterrole=test-clusterrole", "--user=test-user"},
	)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen clusterrolebinding to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
