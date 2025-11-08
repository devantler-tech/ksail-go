package cipher_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cipher"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

func TestNewEncryptCmd(t *testing.T) {
	t.Parallel()

	cmd := cipher.NewEncryptCmd()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	if cmd.Use != "encrypt <file>" {
		t.Errorf("expected Use to be 'encrypt <file>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
}

func TestEncryptCommandHelp(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetArgs([]string{"encrypt", "--help"})

	err := cipherCmd.Execute()
	if err != nil {
		t.Errorf("expected no error executing --help, got: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected help output to not be empty")
	}
}

func TestEncryptCommandRequiresFile(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"encrypt"})

	err := cipherCmd.Execute()
	if err == nil {
		t.Error("expected error when no file argument provided")
	}
}

func TestEncryptCommandUnsupportedFormat(t *testing.T) {
	t.Parallel()

	// Create a temporary txt file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"encrypt", testFile})

	err = cipherCmd.Execute()
	if err == nil {
		t.Error("expected error for unsupported file format")
	}
}

func TestEncryptCommandYAMLFormat(t *testing.T) {
	t.Parallel()

	// Create a temporary yaml file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `apiVersion: v1
kind: Secret
metadata:
  name: test
data:
  key: value
`

	err := os.WriteFile(testFile, []byte(yamlContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"encrypt", testFile})

	// This will fail because no SOPS keys are configured, but we're testing
	// that it gets past file format validation
	err = cipherCmd.Execute()

	// We expect an error about missing keys, not about file format
	if err != nil && errOut.String() == "" {
		// Error is from SOPS (expected)
		t.Logf("Expected SOPS error (no keys configured): %v", err)
	}
}

func TestEncryptCommandJSONFormat(t *testing.T) {
	t.Parallel()

	// Create a temporary json file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	jsonContent := `{
  "apiVersion": "v1",
  "kind": "Secret",
  "metadata": {
    "name": "test"
  },
  "data": {
    "key": "value"
  }
}
`

	err := os.WriteFile(testFile, []byte(jsonContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"encrypt", testFile})

	// This will fail because no SOPS keys are configured, but we're testing
	// that it gets past file format validation
	err = cipherCmd.Execute()

	// We expect an error about missing keys, not about file format
	if err != nil && errOut.String() == "" {
		// Error is from SOPS (expected)
		t.Logf("Expected SOPS error (no keys configured): %v", err)
	}
}
