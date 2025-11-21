package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenPriorityClass tests generating a priorityclass manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenPriorityClass(t *testing.T) {
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
