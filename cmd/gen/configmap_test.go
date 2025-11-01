package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenConfigMap tests generating a configmap manifest.
func TestGenConfigMap(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewConfigMapCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-config", "--from-literal", "APP_ENV=production", "--from-literal", "DEBUG=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen configmap to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
