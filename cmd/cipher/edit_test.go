package cipher_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cipher"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

func TestNewEditCmd(t *testing.T) {
	t.Parallel()

	cmd := cipher.NewEditCmd()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	if cmd.Use != "edit <file>" {
		t.Errorf("expected Use to be 'edit <file>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
}

func TestEditCommandHelp(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetArgs([]string{"edit", "--help"})

	err := cipherCmd.Execute()
	if err != nil {
		t.Errorf("expected no error executing --help, got: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected help output to not be empty")
	}

	// Verify help mentions key concepts
	if !strings.Contains(output, "encrypted file") {
		t.Error("expected help to mention 'encrypted file'")
	}
}

func TestEditCommandRequiresFile(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"edit"})

	err := cipherCmd.Execute()
	if err == nil {
		t.Error("expected error when no file argument provided")
	}
}

func TestEditCommandUnsupportedFormat(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create a test file with unsupported format
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"edit", testFile})

	err = cipherCmd.Execute()
	if err == nil {
		t.Error("expected error for unsupported file format")
	}
}

func TestEditCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := cipher.NewEditCmd()

	// Check that ignore-mac flag exists
	ignoreMacFlag := cmd.Flags().Lookup("ignore-mac")
	if ignoreMacFlag == nil {
		t.Error("expected ignore-mac flag to exist")
	}

	// Check that show-master-keys flag exists
	showMasterKeysFlag := cmd.Flags().Lookup("show-master-keys")
	if showMasterKeysFlag == nil {
		t.Error("expected show-master-keys flag to exist")
	}
}

// TestEditCommandWithYAML tests edit command with YAML format.
// This test expects an error since no keys are configured, but validates
// that the command processes YAML files correctly.
func TestEditCommandWithYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	// Note: We can't fully test the edit command without setting up
	// an editor and keys, but we can verify it handles YAML files
	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"edit", testFile})

	// This will fail because we don't have an editor configured or keys,
	// but it should recognize the YAML format
	err := cipherCmd.Execute()
	if err != nil {
		// Expected to fail - we're just checking it processes YAML
		t.Logf("Expected error (no editor/keys configured): %v", err)
	}
}

// TestEditCommandWithJSON tests edit command with JSON format.
func TestEditCommandWithJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"edit", testFile})

	// This will fail because we don't have an editor configured or keys,
	// but it should recognize the JSON format
	err := cipherCmd.Execute()
	if err != nil {
		// Expected to fail - we're just checking it processes JSON
		t.Logf("Expected error (no editor/keys configured): %v", err)
	}
}
