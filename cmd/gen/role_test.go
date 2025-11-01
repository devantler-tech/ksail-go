package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenRole tests generating a role manifest.
func TestGenRole(t *testing.T) {
	t.Parallel()

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
