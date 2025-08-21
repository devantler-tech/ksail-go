package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewDownCmd(t *testing.T) {
	// Arrange
	t.Parallel()

	// Act
	cmd := cmd.NewDownCmd()

	// Assert
	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "down" {
		t.Fatalf("expected Use to be 'down', got %q", cmd.Use)
	}

	if cmd.Short != "Stop and remove the Kubernetes cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestDownCmd_Execute(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewDownCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Cluster stopped and removed successfully (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}
