package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStopCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewStopCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "stop" {
		t.Fatalf("expected Use to be 'stop', got %q", cmd.Use)
	}

	if cmd.Short != "Stop the Kubernetes cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestStopCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStopCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Cluster stopped successfully (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}
