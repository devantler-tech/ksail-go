package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

func TestNewStopCmd(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewStopCmd(runtimeContainer)

	if cmd.Use != "stop" {
		t.Fatalf("expected Use to be 'stop', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("expected Short description to be set")
	}

	if cmd.RunE == nil {
		t.Fatal("expected RunE to be set")
	}

	var out bytes.Buffer
	cmd.SetOut(&out)
}
