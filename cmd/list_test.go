package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewListCmd(test *testing.T) {
	// Arrange
	test.Parallel()

	// Act
	cmd := cmd.NewListCmd()

	// Assert
	if cmd == nil {
		test.Fatal("expected command to be created")
	}

	if cmd.Use != "list" {
		test.Fatalf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short != "List Kubernetes clusters" {
		test.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestListCmd_Execute_Default(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		test.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Listing running clusters (stub implementation)\n"

	if got != expected {
		test.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestListCmd_Execute_All(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--all"})

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		test.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Listing all clusters (stub implementation)\n"

	if got != expected {
		test.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestListCmd_Flags(test *testing.T) {
	// Arrange
	test.Parallel()

	cmd := cmd.NewListCmd()

	// Act & Assert
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		test.Fatal("expected all flag to exist")
	}

	if allFlag.DefValue != "false" {
		test.Fatalf("expected all flag default to be 'false', got %q", allFlag.DefValue)
	}
}