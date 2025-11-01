package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenCronJob tests generating a cronjob manifest.
func TestGenCronJob(t *testing.T) {
	t.Parallel()

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
