package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenConfigMap tests generating a configmap manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenConfigMap(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewConfigMapCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs(
		[]string{
			"test-config",
			"--from-literal=APP_ENV=production",
			"--from-literal=DEBUG=false",
		},
	)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen configmap to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
