package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestFileWriter_TryWrite_EmptyOutput(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "test content"

	// Act
	result, err := writer.TryWrite(content, "", false)

	// Assert
	assert.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite()")
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
	assert.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite()")

	// Verify file was written
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	assert.NoError(t, err, "ReadFile()")
	assert.Equal(t, content, string(writtenContent), "written file content")
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
	assert.NoError(t, err, "WriteFile() setup")

	// Act
	result, err := writer.TryWrite(newContent, outputPath, false)

	// Assert
	assert.NoError(t, err, "TryWrite()")
	assert.Equal(t, newContent, result, "TryWrite()")

	// Verify file was NOT overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	assert.NoError(t, err, "ReadFile()")
	assert.Equal(t, originalContent, string(writtenContent), "file content (should not be overwritten)")
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
	assert.NoError(t, err, "WriteFile() setup")

	// Act
	result, err := writer.TryWrite(newContent, outputPath, true)

	// Assert
	assert.NoError(t, err, "TryWrite()")
	assert.Equal(t, newContent, result, "TryWrite()")

	// Verify file was overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	assert.NoError(t, err, "ReadFile()")
	assert.Equal(t, newContent, string(writtenContent), "file content (should be overwritten)")
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
	assert.NoError(t, err, "Mkdir() setup")

	// Act
	result, err := writer.TryWrite(content, outputPath, false)

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to check file", "TryWrite() stat failure")
	assert.Equal(t, "", result, "TryWrite() result on error")
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

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to write file", "TryWrite() write failure")
	assert.Equal(t, "", result, "TryWrite() result on error")
}



