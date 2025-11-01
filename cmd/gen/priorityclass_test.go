package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenPriorityClass tests generating a priorityclass manifest.
func TestGenPriorityClass(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewPriorityClassCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-priority", "--value=1000", "--description=Test priority class"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen priorityclass to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
