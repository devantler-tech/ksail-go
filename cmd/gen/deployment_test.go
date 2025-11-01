package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenDeployment tests generating a deployment manifest.
func TestGenDeployment(t *testing.T) {
	t.Parallel()

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
