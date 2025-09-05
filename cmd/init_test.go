package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewInitCmd(t *testing.T) {
	// Arrange
	t.Parallel()

	// Act
	cmd := cmd.NewInitCmd()

	// Assert
	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "init" {
		t.Fatalf("expected Use to be 'init', got %q", cmd.Use)
	}

	if cmd.Short != "Initialize a new KSail project" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestInitCmd_Execute(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewInitCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Project initialized successfully (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestInitCmd_Flags(t *testing.T) {
	// Arrange
	t.Parallel()

	cmd := cmd.NewInitCmd()

	// Act & Assert
	distributionFlag := cmd.Flags().Lookup("distribution")
	if distributionFlag == nil {
		t.Fatal("expected distribution flag to exist")
	}

	if distributionFlag.DefValue != "Kind" {
		t.Fatalf("expected distribution default to be 'Kind', got %q", distributionFlag.DefValue)
	}
}
