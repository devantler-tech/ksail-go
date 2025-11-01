package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenDeployment tests generating a deployment manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenDeployment(t *testing.T) {

	rt := newTestRuntime()
	cmd := NewDeploymentCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-deployment", "--image=nginx:1.21", "--replicas=3"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen deployment to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
