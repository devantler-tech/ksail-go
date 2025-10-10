package sops_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/sops"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCreateCipherCommand(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	if cmd.Use != "cipher" {
		t.Errorf("expected Use to be 'cipher', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if !cmd.DisableFlagParsing {
		t.Error("expected DisableFlagParsing to be true")
	}
}

func TestCipherCommandHasLongDescription(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	// The command should have a long description that mentions the sops dependency
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify the description mentions the sops binary dependency
	if !strings.Contains(cmd.Long, "sops") {
		t.Error("expected Long description to mention 'sops' binary")
	}

	if !strings.Contains(cmd.Long, "Dependencies") {
		t.Error("expected Long description to mention Dependencies")
	}
}

func TestCipherCommand_SopsNotFound(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Create a command with modified PATH that doesn't include sops
	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var outBuf, errBuf bytes.Buffer

	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())

	// Set PATH to empty directory to ensure sops is not found
	tempDir := t.TempDir()

	t.Setenv("PATH", tempDir)

	// Execute command with args
	cmd.SetArgs([]string{"--version"})
	err := cmd.Execute()

	// Should return error when sops not found
	if err == nil {
		t.Fatal("expected error when sops binary not found")
	}

	if !strings.Contains(err.Error(), "sops binary not found") {
		t.Errorf("expected error message about sops not found, got: %v", err)
	}

	if !strings.Contains(err.Error(), "https://github.com/getsops/sops") {
		t.Errorf("expected error message to include installation link, got: %v", err)
	}
}

func TestCipherCommand_WithMockSops(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Create a temporary directory for our mock sops binary
	tempDir := t.TempDir()
	mockSopsPath := filepath.Join(tempDir, "sops")

	// Create a simple shell script that acts as a mock sops
	mockScript := `#!/bin/sh
echo "SOPS mock version 1.0.0"
exit 0
`
	//nolint:gosec // This is a test file with safe permissions for executables
	err := os.WriteFile(mockSopsPath, []byte(mockScript), 0o755)
	if err != nil {
		t.Fatalf("failed to create mock sops: %v", err)
	}

	// Save original PATH and modify it to include our mock
	originalPath := os.Getenv("PATH")

	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)

	// Create and execute command
	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var outBuf, errBuf bytes.Buffer

	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--version"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing mock sops: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "SOPS mock version") {
		t.Errorf("expected output from mock sops, got: %q", output)
	}
}

func TestCipherCommand_SopsExecutionFailure(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Create a temporary directory for our failing mock sops binary
	tempDir := t.TempDir()
	mockSopsPath := filepath.Join(tempDir, "sops")

	// Create a script that exits with error
	mockScript := `#!/bin/sh
echo "Error: invalid command" >&2
exit 1
`
	//nolint:gosec // This is a test file with safe permissions for executables
	err := os.WriteFile(mockSopsPath, []byte(mockScript), 0o755)
	if err != nil {
		t.Fatalf("failed to create mock sops: %v", err)
	}

	// Save original PATH and modify it to include our mock
	originalPath := os.Getenv("PATH")

	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)

	// Create and execute command
	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var outBuf, errBuf bytes.Buffer

	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--invalid-flag"})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error when sops execution fails")
	}

	if !strings.Contains(err.Error(), "sops execution failed") {
		t.Errorf("expected error message about sops execution failure, got: %v", err)
	}
}

func TestCipherCommand_NilContext(t *testing.T) {
	t.Parallel()

	// Skip if sops is not available
	_, err := exec.LookPath("sops")
	if err != nil {
		t.Skip("sops binary not available for testing")
	}

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var outBuf, errBuf bytes.Buffer

	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	// Don't set context - it should default to Background
	cmd.SetArgs([]string{"--version"})

	execErr := cmd.Execute()
	// We don't care if it succeeds or fails, just that it doesn't panic
	// with nil context
	_ = execErr
}

//nolint:paralleltest // Cannot use t.Parallel() with system dependencies
func TestCipherCommand_RealSops(t *testing.T) {
	// Cannot use t.Parallel() with system dependencies

	// Skip if sops is not available
	sopsPath, err := exec.LookPath("sops")
	if err != nil {
		t.Skip("sops binary not available for testing")
	}

	t.Logf("Testing with sops at: %s", sopsPath)

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var outBuf, errBuf bytes.Buffer

	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--version"})

	err = cmd.Execute()
	if err != nil {
		t.Logf("sops --version output: %s", outBuf.String())
		t.Logf("sops --version error: %s", errBuf.String())
		t.Fatalf("unexpected error executing real sops --version: %v", err)
	}

	output := outBuf.String()
	if output == "" {
		t.Error("expected non-empty output from sops --version")
	}
}

func TestCipherCommand_CommandProperties(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	// Test command properties
	tests := []struct {
		name     string
		check    func() bool
		expected bool
		message  string
	}{
		{
			name:     "DisableFlagParsing",
			check:    func() bool { return cmd.DisableFlagParsing },
			expected: true,
			message:  "DisableFlagParsing should be true",
		},
		{
			name:     "DisableFlagsInUseLine",
			check:    func() bool { return cmd.DisableFlagsInUseLine },
			expected: true,
			message:  "DisableFlagsInUseLine should be true",
		},
		{
			name:     "SilenceUsage",
			check:    func() bool { return cmd.SilenceUsage },
			expected: true,
			message:  "SilenceUsage should be true",
		},
		{
			name:     "HasRunE",
			check:    func() bool { return cmd.RunE != nil },
			expected: true,
			message:  "RunE should be set",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.check()
			if result != testCase.expected {
				t.Errorf("%s: expected %v, got %v", testCase.message, testCase.expected, result)
			}
		})
	}
}

func TestCipherCommand_OutputStreams(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Create a temporary directory for our mock sops binary
	tempDir := t.TempDir()
	mockSopsPath := filepath.Join(tempDir, "sops")

	// Create a script that writes to both stdout and stderr
	const stderrFd = 2

	mockScript := fmt.Sprintf(`#!/bin/sh
echo "stdout output"
echo "stderr output" >&%d
exit 0
`, stderrFd)

	//nolint:gosec // This is a test file with safe permissions for executables
	err := os.WriteFile(mockSopsPath, []byte(mockScript), 0o755)
	if err != nil {
		t.Fatalf("failed to create mock sops: %v", err)
	}

	// Save original PATH and modify it to include our mock
	originalPath := os.Getenv("PATH")

	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)

	// Create and execute command
	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var outBuf, errBuf bytes.Buffer

	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify stdout was captured
	if !strings.Contains(outBuf.String(), "stdout output") {
		t.Errorf("expected stdout to be captured, got: %q", outBuf.String())
	}

	// Verify stderr was captured
	if !strings.Contains(errBuf.String(), "stderr output") {
		t.Errorf("expected stderr to be captured, got: %q", errBuf.String())
	}
}
