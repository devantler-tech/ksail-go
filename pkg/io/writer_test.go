package io_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryWrite_WithBuffer(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "test content for buffer"
	buffer := &bytes.Buffer{}

	// Act
	result, err := ioutils.TryWrite(content, buffer)

	// Assert
	require.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite() result")
	assert.Equal(t, content, buffer.String(), "buffer content")
}

func TestTryWrite_WithStringWriter(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "test content for string writer"
	stringBuilder := &strings.Builder{}

	// Act
	result, err := ioutils.TryWrite(content, stringBuilder)

	// Assert
	require.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite() result")
	assert.Equal(t, content, stringBuilder.String(), "string builder content")
}

func TestTryWrite_WithFailingWriter(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "test content"
	failingWriter := &failingWriter{}

	// Act
	result, err := ioutils.TryWrite(content, failingWriter)

	// Assert
	require.Error(t, err, "TryWrite()")
	assert.Contains(t, err.Error(), "failed to write content", "error message")
	assert.Empty(t, result, "TryWrite() result on error")
}

// failingWriter always returns an error on Write.
type failingWriter struct{}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestTryWriteFile_EmptyOutput(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "test content"

	// Act
	result, err := ioutils.TryWriteFile(content, "", false)

	// Assert
	require.Error(t, err, "TryWriteFile()")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestTryWriteFile_NewFile(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "new file content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.txt")

	// Act
	result, err := ioutils.TryWriteFile(content, outputPath, false)

	// Assert
	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, content, result, "TryWriteFile()")

	// Verify file was written
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, content, string(writtenContent), "written file content")
}

func TestTryWriteFile_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	originalContent := "original content"
	newContent := "new content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0600)
	require.NoError(t, err, "WriteFile() setup")

	// Act
	result, err := ioutils.TryWriteFile(newContent, outputPath, false)

	// Assert
	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, newContent, result, "TryWriteFile()")

	// Verify file was NOT overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, originalContent, string(writtenContent), "file content (should not be overwritten)")
}

func TestTryWriteFile_ExistingFile_Force(t *testing.T) {
	t.Parallel()

	// Arrange
	originalContent := "original content"
	newContent := "new content forced"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing-force.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0600)
	require.NoError(t, err, "WriteFile() setup")

	// Act
	result, err := ioutils.TryWriteFile(newContent, outputPath, true)

	// Assert
	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, newContent, result, "TryWriteFile()")

	// Verify file was overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, newContent, string(writtenContent), "file content (should be overwritten)")
}

func TestTryWriteFile_StatError(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "content for stat error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "restricted", "file.txt")

	// Create a directory with no permissions to simulate stat error
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000)
	require.NoError(t, err, "Mkdir() setup")

	// Act
	result, err := ioutils.TryWriteFile(content, outputPath, false)

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to check file", "TryWriteFile() stat failure")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestTryWriteFile_WriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "content for write error test"

	// Use a path that cannot be written to (directory that doesn't exist)
	invalidPath := "/invalid/nonexistent/deeply/nested/path/file.txt"

	// Act
	result, err := ioutils.TryWriteFile(content, invalidPath, false)

	// Assert - expect error containing specific message about directory creation failure
	testutils.AssertErrContains(t, err, "failed to create directory", "TryWriteFile() directory creation failure")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestGetWriter_Quiet(t *testing.T) {
	t.Parallel()

	// Act
	writer := ioutils.GetWriter(true)

	// Assert
	if writer != io.Discard {
		t.Errorf("expected io.Discard for quiet=true, got %T", writer)
	}
}

func TestGetWriter_NotQuiet(t *testing.T) {
	t.Parallel()

	// Act
	writer := ioutils.GetWriter(false)

	// Assert
	if writer != os.Stdout {
		t.Errorf("expected os.Stdout for quiet=false, got %T", writer)
	}
}
