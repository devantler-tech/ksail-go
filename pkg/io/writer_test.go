package io_test

import (
	"os"
	"path/filepath"
	"testing"

	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
)

func TestFileWriter_TryWrite_EmptyOutput(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "test content"

	// Act
	result, err := writer.TryWrite(content, "", false)

	// Assert
	if err != nil {
		t.Fatalf("TryWrite() unexpected error: %v", err)
	}

	if result != content {
		t.Fatalf("TryWrite() = %q, want %q", result, content)
	}
}

func TestFileWriter_TryWrite_NewFile(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "new file content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.txt")

	// Act
	result, err := writer.TryWrite(content, outputPath, false)

	// Assert
	if err != nil {
		t.Fatalf("TryWrite() unexpected error: %v", err)
	}

	if result != content {
		t.Fatalf("TryWrite() = %q, want %q", result, content)
	}

	// Verify file was written
	writtenContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(writtenContent) != content {
		t.Fatalf("written file content = %q, want %q", string(writtenContent), content)
	}
}

func TestFileWriter_TryWrite_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	originalContent := "original content"
	newContent := "new content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0600)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	// Act
	result, err := writer.TryWrite(newContent, outputPath, false)

	// Assert
	if err != nil {
		t.Fatalf("TryWrite() unexpected error: %v", err)
	}

	if result != newContent {
		t.Fatalf("TryWrite() = %q, want %q", result, newContent)
	}

	// Verify file was NOT overwritten
	writtenContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read existing file: %v", err)
	}

	if string(writtenContent) != originalContent {
		t.Fatalf("file content = %q, want %q (should not be overwritten)", string(writtenContent), originalContent)
	}
}

func TestFileWriter_TryWrite_ExistingFile_Force(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	originalContent := "original content"
	newContent := "new content forced"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing-force.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0600)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	// Act
	result, err := writer.TryWrite(newContent, outputPath, true)

	// Assert
	if err != nil {
		t.Fatalf("TryWrite() unexpected error: %v", err)
	}

	if result != newContent {
		t.Fatalf("TryWrite() = %q, want %q", result, newContent)
	}

	// Verify file was overwritten
	writtenContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read overwritten file: %v", err)
	}

	if string(writtenContent) != newContent {
		t.Fatalf("file content = %q, want %q (should be overwritten)", string(writtenContent), newContent)
	}
}

func TestFileWriter_TryWrite_StatError(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "content for stat error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "restricted", "file.txt")

	// Create a directory with no permissions to simulate stat error
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000)

	if err != nil {
		t.Fatalf("failed to create restricted directory: %v", err)
	}

	// Cleanup with proper permissions
	defer func() {
		_ = os.Chmod(restrictedDir, 0755)
		_ = os.RemoveAll(restrictedDir)
	}()

	// Act
	result, err := writer.TryWrite(content, outputPath, false)

	// Assert
	if err == nil {
		t.Fatal("TryWrite() expected error for stat failure, got nil")
	}

	if result != "" {
		t.Fatalf("TryWrite() = %q, want empty string on error", result)
	}

	if !containsString(err.Error(), "failed to check file") {
		t.Fatalf("TryWrite() error = %q, want to contain 'failed to check file'", err.Error())
	}
}

func TestFileWriter_TryWrite_WriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "content for write error test"

	// Use a path that cannot be written to (directory that doesn't exist)
	invalidPath := "/invalid/nonexistent/deeply/nested/path/file.txt"

	// Act
	result, err := writer.TryWrite(content, invalidPath, false)

	// Assert
	if err == nil {
		t.Fatal("TryWrite() expected error for write failure, got nil")
	}

	if result != "" {
		t.Fatalf("TryWrite() = %q, want empty string on error", result)
	}

	if !containsString(err.Error(), "failed to write file") {
		t.Fatalf("TryWrite() error = %q, want to contain 'failed to write file'", err.Error())
	}
}

// Helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

