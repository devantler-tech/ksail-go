package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenJob tests generating a job manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenJob(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewJobCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-job", "--image=busybox:latest"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen job to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
