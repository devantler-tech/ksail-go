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
}

func TestDecryptCommand_WithValidKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file with encrypted data
	tempDir := t.TempDir()
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")
	testData := "Test secret"

	// Generate a test key and encrypt data using SOPS
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	sopsCipher := aes.NewCipher()
	encrypted, err := sopsCipher.Encrypt(testData, key, "")
	if err != nil {
		t.Fatalf("failed to encrypt test data: %v", err)
	}

	err = os.WriteFile(encryptedFile, []byte(encrypted), 0o600)
	if err != nil {
		t.Fatalf("failed to create encrypted file: %v", err)
	}

	// Decrypt using our command
	encodedKey := base64.StdEncoding.EncodeToString(key)
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, encryptedFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("decrypt command failed: %v", err)
	}

	output := out.String()

	// Verify output contains decrypted data
	if !strings.Contains(output, testData) {
		t.Errorf("expected output to contain %q, got: %q", testData, output)
	}
}

func TestDecryptCommand_WithOutputFile(t *testing.T) {
	t.Parallel()

	// Create a temporary file with encrypted data
	tempDir := t.TempDir()
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")
	outputFile := filepath.Join(tempDir, "decrypted.txt")
	testData := "Secret data"

	// Generate a test key and encrypt data
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	sopsCipher := aes.NewCipher()
	encrypted, err := sopsCipher.Encrypt(testData, key, "")
	if err != nil {
		t.Fatalf("failed to encrypt test data: %v", err)
	}

	err = os.WriteFile(encryptedFile, []byte(encrypted), 0o600)
	if err != nil {
		t.Fatalf("failed to create encrypted file: %v", err)
	}

	// Decrypt with output file
	encodedKey := base64.StdEncoding.EncodeToString(key)
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, "--output", outputFile, encryptedFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("decrypt command failed: %v", err)
	}

	// Verify output file was created
	_, statErr := os.Stat(outputFile)
	if os.IsNotExist(statErr) {
		t.Error("expected output file to be created")
	}

	// Verify output file contains decrypted data
	decryptedData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(decryptedData) != testData {
		t.Errorf("expected output file to contain %q, got %q", testData, string(decryptedData))
	}

	// Verify command output mentions the file
	if !strings.Contains(out.String(), outputFile) {
		t.Error("expected command output to mention output file")
	}
}

func TestDecryptCommand_MissingKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file
	tempDir := t.TempDir()
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")

	err := os.WriteFile(encryptedFile, []byte("ENC[AES256_GCM,data:test,iv:test,tag:test,type:str]"), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Try to decrypt without providing key
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{encryptedFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error when key is not provided")
	}
}

func TestDecryptCommand_InvalidKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file with encrypted data
	tempDir := t.TempDir()
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")

	// Use SOPS to create valid encrypted data
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	sopsCipher := aes.NewCipher()
	encrypted, err := sopsCipher.Encrypt("test", key, "")
	if err != nil {
		t.Fatalf("failed to encrypt test data: %v", err)
	}

	err = os.WriteFile(encryptedFile, []byte(encrypted), 0o600)
	if err != nil {
		t.Fatalf("failed to create encrypted file: %v", err)
	}

	// Test with invalid base64 key
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", "invalid-base64!", encryptedFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error with invalid base64 key")
	}

	// Test with wrong length key
	shortKey := make([]byte, 16)

	_, err = rand.Read(shortKey)
	if err != nil {
		t.Fatalf("failed to generate short key: %v", err)
	}

	encodedShortKey := base64.StdEncoding.EncodeToString(shortKey)

	cmd = cipher.NewDecryptCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedShortKey, encryptedFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error with wrong length key")
	}

	if !strings.Contains(err.Error(), "32 bytes") {
		t.Errorf("expected error message about 32 bytes, got: %v", err)
	}
}

func TestDecryptCommand_WrongKey(t *testing.T) {
	t.Parallel()

	// Create a temporary file with encrypted data
	tempDir := t.TempDir()
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")

	// Encrypt with one key
	key1 := make([]byte, 32)
	_, err := rand.Read(key1)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	sopsCipher := aes.NewCipher()
	encrypted, err := sopsCipher.Encrypt("secret", key1, "")
	if err != nil {
		t.Fatalf("failed to encrypt test data: %v", err)
	}

	err = os.WriteFile(encryptedFile, []byte(encrypted), 0o600)
	if err != nil {
		t.Fatalf("failed to create encrypted file: %v", err)
	}

	// Try to decrypt with different key
	key2 := make([]byte, 32)
	_, err = rand.Read(key2)
	if err != nil {
		t.Fatalf("failed to generate second test key: %v", err)
	}

	encodedKey2 := base64.StdEncoding.EncodeToString(key2)
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey2, encryptedFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}

	if !strings.Contains(err.Error(), "failed to decrypt") {
		t.Errorf("expected error message about decryption failure, got: %v", err)
	}
}

func TestDecryptCommand_InvalidEncryptedData(t *testing.T) {
	t.Parallel()

	// Create a temporary file with invalid encrypted data
	tempDir := t.TempDir()
	encryptedFile := filepath.Join(tempDir, "encrypted.enc")

	err := os.WriteFile(encryptedFile, []byte("Not valid SOPS encrypted data"), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Try to decrypt
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, encryptedFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error when encrypted data is invalid")
	}
}

func TestDecryptCommand_NonExistentFile(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	cmd := cipher.NewDecryptCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--key", encodedKey, "/nonexistent/file.enc"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error when input file does not exist")
	}
}
