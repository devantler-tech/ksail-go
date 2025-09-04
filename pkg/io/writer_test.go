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

func TestFileWriter_TryWrite_WithBuffer(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "test content for buffer"
	buffer := &bytes.Buffer{}

	// Act
	result, err := writer.TryWrite(content, buffer)

	// Assert
	require.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite() result")
	assert.Equal(t, content, buffer.String(), "buffer content")
}

func TestFileWriter_TryWrite_WithStringWriter(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "test content for string writer"
	stringBuilder := &strings.Builder{}

	// Act
	result, err := writer.TryWrite(content, stringBuilder)

	// Assert
	require.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite() result")
	assert.Equal(t, content, stringBuilder.String(), "string builder content")
}

func TestFileWriter_TryWrite_WithFailingWriter(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "test content"
	failingWriter := &failingWriter{}

	// Act
	result, err := writer.TryWrite(content, failingWriter)

	// Assert
	require.Error(t, err, "TryWrite()")
	assert.Contains(t, err.Error(), "failed to write content", "error message")
	assert.Empty(t, result, "TryWrite() result on error")
}

// failingWriter always returns an error on Write
type failingWriter struct{}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestFileWriter_TryWriteFile_EmptyOutput(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "test content"

	// Act
	result, err := writer.TryWriteFile(content, "", false)

	// Assert
	require.Error(t, err, "TryWriteFile()")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestFileWriter_TryWriteFile_NewFile(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "new file content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.txt")

	// Act
	result, err := writer.TryWriteFile(content, outputPath, false)

	// Assert
	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, content, result, "TryWriteFile()")

	// Verify file was written
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, content, string(writtenContent), "written file content")
}

func TestFileWriter_TryWriteFile_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	originalContent := "original content"
	newContent := "new content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0600)
	require.NoError(t, err, "WriteFile() setup")

	// Act
	result, err := writer.TryWriteFile(newContent, outputPath, false)

	// Assert
	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, newContent, result, "TryWriteFile()")

	// Verify file was NOT overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, originalContent, string(writtenContent), "file content (should not be overwritten)")
}

func TestFileWriter_TryWriteFile_ExistingFile_Force(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	originalContent := "original content"
	newContent := "new content forced"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing-force.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0600)
	require.NoError(t, err, "WriteFile() setup")

	// Act
	result, err := writer.TryWriteFile(newContent, outputPath, true)

	// Assert
	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, newContent, result, "TryWriteFile()")

	// Verify file was overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, newContent, string(writtenContent), "file content (should be overwritten)")
}

func TestFileWriter_TryWriteFile_StatError(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "content for stat error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "restricted", "file.txt")

	// Create a directory with no permissions to simulate stat error
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000)
	require.NoError(t, err, "Mkdir() setup")

	// Act
	result, err := writer.TryWriteFile(content, outputPath, false)

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to check file", "TryWriteFile() stat failure")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestFileWriter_TryWriteFile_WriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	writer := ioutils.FileWriter{}
	content := "content for write error test"

	// Use a path that cannot be written to (directory that doesn't exist)
	invalidPath := "/invalid/nonexistent/deeply/nested/path/file.txt"

	// Act
	result, err := writer.TryWriteFile(content, invalidPath, false)

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to write file", "TryWriteFile() write failure")
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
