package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStopCmd(test *testing.T) {
	// Arrange
	test.Parallel()

	// Act
	cmd := cmd.NewStopCmd()

	// Assert
	if cmd == nil {
		test.Fatal("expected command to be created")
	}

	if cmd.Use != "stop" {
		test.Fatalf("expected Use to be 'stop', got %q", cmd.Use)
	}

	if cmd.Short != "Stop the Kubernetes cluster" {
		test.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestStopCmd_Execute(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStopCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		test.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Cluster stopped successfully (stub implementation)\n"

	if got != expected {
		test.Fatalf("expected output %q, got %q", expected, got)
	}
}