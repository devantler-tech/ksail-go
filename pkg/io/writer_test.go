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

const (
	testContent     = "test content"
	originalContent = "original content"
)

func TestTryWriteWithBuffer(t *testing.T) {
	t.Parallel()

	content := "test content for buffer"
	buffer := &bytes.Buffer{}

	result, err := ioutils.TryWrite(content, buffer)

	require.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite() result")
	assert.Equal(t, content, buffer.String(), "buffer content")
}

func TestTryWriteWithStringWriter(t *testing.T) {
	t.Parallel()

	content := "test content for string writer"
	stringBuilder := &strings.Builder{}

	result, err := ioutils.TryWrite(content, stringBuilder)

	require.NoError(t, err, "TryWrite()")
	assert.Equal(t, content, result, "TryWrite() result")
	assert.Equal(t, content, stringBuilder.String(), "string builder content")
}

func TestTryWriteWithFailingWriter(t *testing.T) {
	t.Parallel()

	content := testContent
	failingWriter := &failingWriter{}

	result, err := ioutils.TryWrite(content, failingWriter)

	require.Error(t, err, "TryWrite()")
	assert.Contains(t, err.Error(), "failed to write content", "error message")
	assert.Empty(t, result, "TryWrite() result on error")
}

// failingWriter always returns an error on Write.
type failingWriter struct{}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestTryWriteFileEmptyOutput(t *testing.T) {
	t.Parallel()

	content := testContent

	result, err := ioutils.TryWriteFile(content, "", false)

	require.Error(t, err, "TryWriteFile()")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestTryWriteFileNewFile(t *testing.T) {
	t.Parallel()

	content := "new file content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.txt")

	result, err := ioutils.TryWriteFile(content, outputPath, false)

	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, content, result, "TryWriteFile()")

	// Verify file was written
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, content, string(writtenContent), "written file content")
}

func TestTryWriteFileExistingFileNoForce(t *testing.T) {
	t.Parallel()

	newContent := "new content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile() setup")

	result, err := ioutils.TryWriteFile(newContent, outputPath, false)

	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, newContent, result, "TryWriteFile()")

	// Verify file was NOT overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(
		t,
		originalContent,
		string(writtenContent),
		"file content (should not be overwritten)",
	)
}

func TestTryWriteFileExistingFileForce(t *testing.T) {
	t.Parallel()

	newContent := "new content forced"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing-force.txt")

	// Create existing file
	err := os.WriteFile(outputPath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile() setup")

	result, err := ioutils.TryWriteFile(newContent, outputPath, true)

	require.NoError(t, err, "TryWriteFile()")
	assert.Equal(t, newContent, result, "TryWriteFile()")

	// Verify file was overwritten
	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, newContent, string(writtenContent), "file content (should be overwritten)")
}

func TestTryWriteFileStatError(t *testing.T) {
	t.Parallel()

	content := "content for stat error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "restricted", "file.txt")

	// Create a directory with no permissions to simulate stat error
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0o000)
	require.NoError(t, err, "Mkdir() setup")

	result, err := ioutils.TryWriteFile(content, outputPath, false)

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to check file", "TryWriteFile() stat failure")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestTryWriteFileWriteError(t *testing.T) {
	t.Parallel()

	content := "content for write error test"

	// Use a path that cannot be written to (directory that doesn't exist)
	invalidPath := "/invalid/nonexistent/deeply/nested/path/file.txt"

	result, err := ioutils.TryWriteFile(content, invalidPath, false)

	// Assert - expect error containing specific message about directory creation failure
	testutils.AssertErrContains(
		t,
		err,
		"failed to create directory",
		"TryWriteFile() directory creation failure",
	)
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestGetWriterQuiet(t *testing.T) {
	t.Parallel()

	writer := ioutils.GetWriter(true)

	if writer != io.Discard {
		t.Errorf("expected io.Discard for quiet=true, got %T", writer)
	}
}

func TestTryWriteFileFileWriteError(t *testing.T) {
	t.Parallel()

	content := "content for file write error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "readonly.txt")

	// Create a directory that exists but make it read-only to cause file write failure
	err := os.WriteFile(outputPath, []byte("existing"), 0o000) // No write permissions
	require.NoError(t, err, "WriteFile() setup")

	result, err := ioutils.TryWriteFile(content, outputPath, true) // force=true to skip stat check

	// Assert - expect error containing specific message about file write failure
	testutils.AssertErrContains(t, err, "failed to write file", "TryWriteFile() file write failure")
	assert.Empty(t, result, "TryWriteFile() result on error")
}

func TestWriteFileSafeEmptyBasePath(t *testing.T) {
	t.Parallel()

	content := testContent
	filePath := "/some/path/file.txt"

	err := ioutils.WriteFileSafe(content, "", filePath, false)

	testutils.AssertErrWrappedContains(
		t,
		err,
		ioutils.ErrBasePath,
		"",
		"WriteFileSafe empty base path",
	)
}

func TestWriteFileSafeEmptyFilePath(t *testing.T) {
	t.Parallel()

	content := testContent
	basePath := t.TempDir()

	err := ioutils.WriteFileSafe(content, basePath, "", false)

	testutils.AssertErrWrappedContains(
		t,
		err,
		ioutils.ErrEmptyOutputPath,
		"",
		"WriteFileSafe empty file path",
	)
}

func TestWriteFileSafePathOutsideBase(t *testing.T) {
	t.Parallel()

	content := testContent
	basePath := t.TempDir()
	outsidePath := "/tmp/outside.txt" // Path clearly outside basePath

	err := ioutils.WriteFileSafe(content, basePath, outsidePath, false)

	testutils.AssertErrWrappedContains(
		t,
		err,
		ioutils.ErrPathOutsideBase,
		"",
		"WriteFileSafe path outside base",
	)
}

func TestWriteFileSafeNewFile(t *testing.T) {
	t.Parallel()

	content := "new file content via WriteFileSafe"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "newfile.txt")

	err := ioutils.WriteFileSafe(content, basePath, filePath, false)

	require.NoError(t, err, "WriteFileSafe")

	// Verify file was written
	writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
	require.NoError(t, err, "ReadFile")
	assert.Equal(t, content, string(writtenContent), "file content")
}

func TestWriteFileSafeExistingFileNoForce(t *testing.T) {
	t.Parallel()

	newContent := "new content"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "existing.txt")

	// Create existing file
	err := os.WriteFile(filePath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile setup")

	err = ioutils.WriteFileSafe(newContent, basePath, filePath, false)

	require.NoError(t, err, "WriteFileSafe")

	// Verify file was NOT overwritten
	writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
	require.NoError(t, err, "ReadFile")
	assert.Equal(
		t,
		originalContent,
		string(writtenContent),
		"file content should not be overwritten",
	)
}

func TestWriteFileSafeExistingFileForce(t *testing.T) {
	t.Parallel()

	newContent := "new content forced"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "existing-force.txt")

	// Create existing file
	err := os.WriteFile(filePath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile setup")

	err = ioutils.WriteFileSafe(newContent, basePath, filePath, true)

	require.NoError(t, err, "WriteFileSafe")

	// Verify file was overwritten
	writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
	require.NoError(t, err, "ReadFile")
	assert.Equal(t, newContent, string(writtenContent), "file content should be overwritten")
}

func TestWriteFileSafeStatError(t *testing.T) {
	t.Parallel()

	content := "content for stat error test"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "restricted", "file.txt")

	// Create a directory with no permissions to simulate stat error
	restrictedDir := filepath.Join(basePath, "restricted")
	err := os.Mkdir(restrictedDir, 0o000)
	require.NoError(t, err, "Mkdir setup")

	err = ioutils.WriteFileSafe(content, basePath, filePath, false)

	// Assert - expect error containing specific message
	testutils.AssertErrContains(t, err, "failed to check file", "WriteFileSafe stat failure")
}

func TestWriteFileSafeDirectoryCreationError(t *testing.T) {
	t.Parallel()

	content := "content for directory creation error test"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "subdir", "file.txt")

	// Create a file with the same name as the directory we want to create
	subdirPath := filepath.Join(basePath, "subdir")
	err := os.WriteFile(subdirPath, []byte("blocking file"), 0o600)
	require.NoError(t, err, "WriteFile setup to block directory creation")

	err = ioutils.WriteFileSafe(
		content,
		basePath,
		filePath,
		true,
	) // Use force=true to skip stat check

	// Assert - expect error containing specific message about directory creation failure
	testutils.AssertErrContains(
		t,
		err,
		"failed to create directory",
		"WriteFileSafe directory creation failure",
	)
}

func TestWriteFileSafeFileWriteError(t *testing.T) {
	t.Parallel()

	content := "content for file write error test"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "readonly.txt")

	// Create a file with no write permissions to cause write failure
	err := os.WriteFile(filePath, []byte("existing"), 0o000) // No write permissions
	require.NoError(t, err, "WriteFile setup")

	err = ioutils.WriteFileSafe(content, basePath, filePath, true) // force=true to skip stat check

	// Assert - expect error containing specific message about file write failure
	testutils.AssertErrContains(t, err, "failed to write file", "WriteFileSafe file write failure")
}

func TestGetWriterNotQuiet(t *testing.T) {
	t.Parallel()

	writer := ioutils.GetWriter(false)

	if writer != os.Stdout {
		t.Errorf("expected os.Stdout for quiet=false, got %T", writer)
	}
}
