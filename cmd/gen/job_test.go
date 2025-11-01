package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenJob tests generating a job manifest.
func TestGenJob(t *testing.T) {
	t.Parallel()

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
