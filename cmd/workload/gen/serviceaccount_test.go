package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenServiceAccount tests generating a serviceaccount manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenServiceAccount(t *testing.T) {
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
