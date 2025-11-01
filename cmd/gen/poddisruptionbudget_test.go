package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenPodDisruptionBudget tests generating a poddisruptionbudget manifest.
func TestGenPodDisruptionBudget(t *testing.T) {
	t.Parallel()

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
