package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenPodDisruptionBudget tests generating a poddisruptionbudget manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenPodDisruptionBudget(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewPodDisruptionBudgetCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-pdb", "--min-available=2", "--selector=app=test"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen poddisruptionbudget to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
