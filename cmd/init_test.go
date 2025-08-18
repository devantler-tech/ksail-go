package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewInitCmd(test *testing.T) {
	// Arrange
	test.Parallel()

	// Act
	cmd := cmd.NewInitCmd()

	// Assert
	if cmd == nil {
		test.Fatal("expected command to be created")
	}

	if cmd.Use != "init" {
		test.Fatalf("expected Use to be 'init', got %q", cmd.Use)
	}

	if cmd.Short != "Initialize a new KSail project" {
		test.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestInitCmd_Execute(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewInitCmd()
	cmd.SetOut(&out)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		test.Fatalf("expected no error, got %v", err)
	}

	got := out.String()

	expected := "✔ Project initialized successfully (stub implementation)\n"

	if got != expected {
		test.Fatalf("expected output %q, got %q", expected, got)
	}
}

func TestInitCmd_Flags(test *testing.T) {
	// Arrange
	test.Parallel()

	cmd := cmd.NewInitCmd()

	// Act & Assert
	containerEngineFlag := cmd.Flags().Lookup("container-engine")
	if containerEngineFlag == nil {
		test.Fatal("expected container-engine flag to exist")
	}

	if containerEngineFlag.DefValue != "Docker" {
		test.Fatalf("expected container-engine default to be 'Docker', got %q", containerEngineFlag.DefValue)
	}

	distributionFlag := cmd.Flags().Lookup("distribution")
	if distributionFlag == nil {
		test.Fatal("expected distribution flag to exist")
	}

	if distributionFlag.DefValue != "Kind" {
		test.Fatalf("expected distribution default to be 'Kind', got %q", distributionFlag.DefValue)
	}
}