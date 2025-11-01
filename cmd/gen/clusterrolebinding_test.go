package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenClusterRoleBinding tests generating a clusterrolebinding manifest.
func TestGenClusterRoleBinding(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewClusterRoleBindingCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-clusterrolebinding", "--clusterrole=test-clusterrole", "--user=test-user"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen clusterrolebinding to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
