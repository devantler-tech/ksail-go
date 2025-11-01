package cipher_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cipher"
	"github.com/getsops/sops/v3/aes"
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

func TestEncryptCommand_WithRandomKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file with test data
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	testData := "Hello, World!"

	err := os.WriteFile(inputFile, []byte(testData), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Execute encrypt command
	cmd := cipher.NewEncryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{inputFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	output := out.String()

	// Verify output contains encrypted data in SOPS format
	if !strings.Contains(output, "ENC[AES256_GCM,") {
		t.Error("expected output to contain SOPS-format encrypted data")
	}

	// Verify output mentions generated key
	if !strings.Contains(output, "Generated key:") {
		t.Error("expected output to mention generated key")
	}
}

func TestEncryptCommand_WithProvidedKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file with test data
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	testData := "Secret message"

	err := os.WriteFile(inputFile, []byte(testData), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Generate a test key
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Execute encrypt command with key
	cmd := cipher.NewEncryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, inputFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	output := out.String()

	// Verify output contains encrypted data
	if !strings.Contains(output, "ENC[AES256_GCM,") {
		t.Error("expected output to contain SOPS-format encrypted data")
	}

	// Should NOT mention generated key when key is provided
	if strings.Contains(output, "Generated key:") {
		t.Error("expected output to NOT mention generated key when key is provided")
	}
}

func TestEncryptCommand_WithOutputFile(t *testing.T) {
	t.Parallel()

	// Create a temporary file with test data
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	outputFile := filepath.Join(tempDir, "output.enc")
	testData := "Test data for encryption"

	err := os.WriteFile(inputFile, []byte(testData), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Generate a test key
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Execute encrypt command with output file
	cmd := cipher.NewEncryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, "--output", outputFile, inputFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	// Verify output file was created
	_, statErr := os.Stat(outputFile)
	if os.IsNotExist(statErr) {
		t.Error("expected output file to be created")
	}

	// Verify output file contains encrypted data
	encryptedData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if !strings.Contains(string(encryptedData), "ENC[AES256_GCM,") {
		t.Error("expected output file to contain SOPS-format encrypted data")
	}

	// Verify command output mentions the file
	if !strings.Contains(out.String(), outputFile) {
		t.Error("expected command output to mention output file")
	}
}

func TestEncryptCommand_InvalidKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file with test data
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")

	err := os.WriteFile(inputFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test with invalid base64 key
	cmd := cipher.NewEncryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", "invalid-base64!", inputFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error with invalid base64 key")
	}

	// Test with wrong length key (16 bytes instead of 32)
	shortKey := make([]byte, 16)

	_, err = rand.Read(shortKey)
	if err != nil {
		t.Fatalf("failed to generate short key: %v", err)
	}

	encodedShortKey := base64.StdEncoding.EncodeToString(shortKey)

	cmd = cipher.NewEncryptCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedShortKey, inputFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error with wrong length key")
	}

	if !strings.Contains(err.Error(), "32 bytes") {
		t.Errorf("expected error message about 32 bytes, got: %v", err)
	}
}

func TestEncryptCommand_NonExistentFile(t *testing.T) {
	t.Parallel()

	cmd := cipher.NewEncryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"/nonexistent/file.txt"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when input file does not exist")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	// Create a temporary file with test data
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")
	decryptedFile := filepath.Join(tempDir, "decrypted.txt")
	testData := "Round trip test data!"

	err := os.WriteFile(inputFile, []byte(testData), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Generate a test key
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Encrypt
	encryptCmd := cipher.NewEncryptCmd()
	var encOut bytes.Buffer
	encryptCmd.SetOut(&encOut)
	encryptCmd.SetArgs([]string{"--key", encodedKey, "--output", encryptedFile, inputFile})

	err = encryptCmd.Execute()
	if err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	// Decrypt
	decryptCmd := cipher.NewDecryptCmd()
	var decOut bytes.Buffer
	decryptCmd.SetOut(&decOut)
	decryptCmd.SetArgs([]string{"--key", encodedKey, "--output", decryptedFile, encryptedFile})

	err = decryptCmd.Execute()
	if err != nil {
		t.Fatalf("decrypt command failed: %v", err)
	}

	// Verify decrypted data matches original
	decryptedData, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if string(decryptedData) != testData {
		t.Errorf("decrypted data does not match original. Expected %q, got %q", testData, string(decryptedData))
	}
}

func TestEncryptCommand_CompatibleWithSOPS(t *testing.T) {
	t.Parallel()

	// Create a temporary file with test data
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	testData := "SOPS compatibility test"

	err := os.WriteFile(inputFile, []byte(testData), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Generate a test key
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Encrypt using our command
	cmd := cipher.NewEncryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, inputFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	// Extract the encrypted string from output
	output := out.String()
	lines := strings.Split(output, "\n")
	var encryptedString string

	for _, line := range lines {
		if strings.HasPrefix(line, "ENC[AES256_GCM,") {
			encryptedString = line

			break
		}
	}

	if encryptedString == "" {
		t.Fatal("could not find encrypted string in output")
	}

	// Verify we can decrypt it using SOPS AES cipher directly
	sopsCipher := aes.NewCipher()
	decrypted, err := sopsCipher.Decrypt(encryptedString, key, "")
	if err != nil {
		t.Fatalf("failed to decrypt with SOPS cipher: %v", err)
	}

	decryptedStr, ok := decrypted.(string)
	if !ok {
		t.Fatalf("unexpected decrypted type: %T", decrypted)
	}

	if decryptedStr != testData {
		t.Errorf("SOPS-decrypted data does not match original. Expected %q, got %q", testData, decryptedStr)
	}
}
