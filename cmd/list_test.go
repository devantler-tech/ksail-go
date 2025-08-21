package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewListCmd(t *testing.T) {
	// Arrange
	t.Parallel()

	// Act
	cmd := cmd.NewListCmd()

	// Assert
	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "list" {
		t.Fatalf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short != "List Kubernetes clusters" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestListCmd_Execute_Default(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Listing running clusters (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestListCmd_Execute_All(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--all"})

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Listing all clusters (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestListCmd_Flags(t *testing.T) {
	// Arrange
	t.Parallel()

	cmd := cmd.NewListCmd()

	// Act & Assert
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("expected all flag to exist")
	}

	if allFlag.DefValue != "false" {
		t.Fatalf("expected all flag default to be 'false', got %q", allFlag.DefValue)
	}
}
