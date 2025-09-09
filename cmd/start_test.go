package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStartCmd(t *testing.T) {
	// Arrange
	t.Parallel()

	// Act
	cmd := cmd.NewStartCmd()

	// Assert
	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "start" {
		t.Fatalf("expected Use to be 'start', got %q", cmd.Use)
	}

	if cmd.Short != "Start a stopped Kubernetes cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestStartCmd_Execute(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStartCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()
	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Cluster started successfully (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}
