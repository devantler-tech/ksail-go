package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenRoleBinding tests generating a rolebinding manifest.
func TestGenRoleBinding(t *testing.T) {
	t.Parallel()

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
