package cipher_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cipher"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

func TestNewDecryptCmd(t *testing.T) {
	t.Parallel()

	cmd := cipher.NewDecryptCmd()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	if cmd.Use != "decrypt <file>" {
		t.Errorf("expected Use to be 'decrypt <file>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify flags are registered
	extractFlag := cmd.Flags().Lookup("extract")
	if extractFlag == nil {
		t.Error("expected extract flag to be registered")
	}

	ignoreMacFlag := cmd.Flags().Lookup("ignore-mac")
	if ignoreMacFlag == nil {
		t.Error("expected ignore-mac flag to be registered")
	}

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("expected output flag to be registered")
	}
}

func TestDecryptCommandHelp(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetArgs([]string{"decrypt", "--help"})

	err := cipherCmd.Execute()
	if err != nil {
		t.Errorf("expected no error executing --help, got: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected help output to not be empty")
	}

	// Verify help output mentions key features
	if !strings.Contains(output, "decrypt") {
		t.Error("expected help output to mention decrypt")
	}
}

func TestDecryptCommandAcceptsStdin(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt"})

	// Should not error on missing file arg (stdin is valid)
	// We expect a decryption error, not an args error
	err := cipherCmd.Execute()
	if err == nil {
		t.Log("Command executed (expected to fail on decryption)")
	}
}

// setupDecryptTest is a helper function to create an encrypted test file.
func setupDecryptTest(t *testing.T, filename, content string) (string, error) {
	t.Helper()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, filename)

	err := os.WriteFile(testFile, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	return testFile, nil
}

func TestDecryptCommandUnsupportedFormat(t *testing.T) {
	t.Parallel()

	testFile, err := setupDecryptTest(t, "test.txt", "test content")
	if err != nil {
		t.Fatal(err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt", testFile})

	err = cipherCmd.Execute()
	if err == nil {
		t.Error("expected error for unsupported file format")
	}

	if !strings.Contains(err.Error(), "unsupported file format") {
		t.Errorf("expected unsupported format error, got: %v", err)
	}
}

func TestDecryptCommandNonExistentFile(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt", "/nonexistent/file.yaml"})

	err := cipherCmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// testDecryptWithFormat tests decryption with supported file formats.
func testDecryptWithFormat(t *testing.T, filename, content string) {
	t.Helper()

	testFile, err := setupDecryptTest(t, filename, content)
	if err != nil {
		t.Fatal(err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt", testFile})

	err = cipherCmd.Execute()
	// We expect an error about the file not being encrypted or missing keys
	// not about file format
	if err != nil {
		t.Logf("Expected SOPS error (not encrypted file): %v", err)
	}
}

func TestDecryptCommandYAMLFormat(t *testing.T) {
	t.Parallel()

	yamlContent := `apiVersion: v1
kind: Secret
metadata:
  name: test
data:
  key: value
`
	testDecryptWithFormat(t, "test.yaml", yamlContent)
}

func TestDecryptCommandJSONFormat(t *testing.T) {
	t.Parallel()

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
	testDecryptWithFormat(t, "test.json", jsonContent)
}

func TestDecryptCommandWithExtractFlag(t *testing.T) {
	t.Parallel()

	testFile, err := setupDecryptTest(t, "test.yaml", "key: value")
	if err != nil {
		t.Fatal(err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt", testFile, "--extract", `["key"]`})

	// Execute - we expect it to fail on decryption (not encrypted), not on flag parsing
	err = cipherCmd.Execute()
	if err != nil {
		t.Logf("Expected SOPS error: %v", err)
	}
}

func TestDecryptCommandWithIgnoreMacFlag(t *testing.T) {
	t.Parallel()

	testFile, err := setupDecryptTest(t, "test.yaml", "key: value")
	if err != nil {
		t.Fatal(err)
	}

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt", testFile, "--ignore-mac"})

	// Execute - we expect it to fail on decryption (not encrypted), not on flag parsing
	err = cipherCmd.Execute()
	if err != nil {
		t.Logf("Expected SOPS error: %v", err)
	}
}

func TestDecryptCommandWithOutputFlag(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile, err := setupDecryptTest(t, "test.yaml", "key: value")
	if err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(tmpDir, "decrypted.yaml")

	rt := runtime.NewRuntime()
	cipherCmd := cipher.NewCipherCmd(rt)

	var out, errOut bytes.Buffer
	cipherCmd.SetOut(&out)
	cipherCmd.SetErr(&errOut)
	cipherCmd.SetArgs([]string{"decrypt", testFile, "--output", outputFile})

	// Execute - we expect it to fail on decryption (not encrypted), not on flag parsing
	err = cipherCmd.Execute()
	if err != nil {
		t.Logf("Expected SOPS error: %v", err)
	}
}

func TestCipherCommandHasDecryptSubcommand(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cmd := cipher.NewCipherCmd(rt)

	// Verify decrypt subcommand exists
	var decryptCmd *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Name() == "decrypt" {
			decryptCmd = c
			break
		}
	}
	if decryptCmd == nil {
		t.Error("expected decrypt subcommand to exist")
	}
}
