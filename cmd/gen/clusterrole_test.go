package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenClusterRole tests generating a clusterrole manifest.
func TestGenClusterRole(t *testing.T) {
	t.Parallel()

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
