package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenCronJob tests generating a cronjob manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenCronJob(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewCronJobCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-cronjob", "--image=busybox:latest", "--schedule=*/5 * * * *"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen cronjob to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
