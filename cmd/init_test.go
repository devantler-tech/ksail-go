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

	expected := "✔ Project initialized successfully with Kind distribution (stub implementation)\n► Cluster name: ksail-default\n► Source directory: k8s\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestInitCmd_Flags(t *testing.T) {
	// Arrange
	t.Parallel()

	cmd := cmd.NewInitCmd()

	// Act & Assert
	distributionFlag := cmd.Flags().Lookup("spec-distribution")
	if distributionFlag == nil {
		t.Fatal("expected spec-distribution flag to exist")
	}

	// Following Viper best practices: CLI flags should not have defaults
	if distributionFlag.DefValue != "" {
		t.Fatalf("expected distribution default to be empty (no CLI defaults), got %q", distributionFlag.DefValue)
	}
}
