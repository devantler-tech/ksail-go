package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenServiceAccount tests generating a serviceaccount manifest.
func TestGenServiceAccount(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewServiceAccountCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-sa"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen serviceaccount to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
