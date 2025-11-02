package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

func newTestRuntime() *runtime.Runtime {
	return runtime.NewRuntime()
}

// TestGenNamespace tests generating a namespace manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenNamespace(t *testing.T) {
	rt := runtime.NewRuntime()
	cmd := NewNamespaceCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-namespace"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen namespace to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
