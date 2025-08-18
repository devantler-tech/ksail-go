package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStartCmd(test *testing.T) {
	// Arrange
	test.Parallel()

	// Act
	cmd := cmd.NewStartCmd()

	// Assert
	if cmd == nil {
		test.Fatal("expected command to be created")
	}

	if cmd.Use != "start" {
		test.Fatalf("expected Use to be 'start', got %q", cmd.Use)
	}

	if cmd.Short != "Start a stopped Kubernetes cluster" {
		test.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestStartCmd_Execute(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStartCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		test.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Cluster started successfully (stub implementation)\n"

	if got != expected {
		test.Fatalf("expected output %q, got %q", expected, got)
	}
}