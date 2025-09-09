package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewUpCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewUpCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "up" {
		t.Fatalf("expected Use to be 'up', got %q", cmd.Use)
	}

	if cmd.Short != "Start the Kubernetes cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestUpCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewUpCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Cluster created and started successfully (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}
