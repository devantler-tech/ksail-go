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

func TestTryWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		content         string
		setupWriter     func() io.Writer
		expectError     bool
		expectedResult  string
		expectedContent string
	}{
		{
			name:    "with buffer",
			content: "test content for buffer",
			setupWriter: func() io.Writer {
				return &bytes.Buffer{}
			},
			expectError:     false,
			expectedResult:  "test content for buffer",
			expectedContent: "test content for buffer",
		},
		{
			name:    "with string writer",
			content: "test content for string writer",
			setupWriter: func() io.Writer {
				return &strings.Builder{}
			},
			expectError:     false,
			expectedResult:  "test content for string writer",
			expectedContent: "test content for string writer",
		},
		{
			name:    "with failing writer",
			content: testContent,
			setupWriter: func() io.Writer {
				return &failingWriter{}
			},
			expectError:    true,
			expectedResult: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			writer := test.setupWriter()
			result, err := ioutils.TryWrite(test.content, writer)

			if test.expectError {
				require.Error(t, err, "TryWrite()")
				assert.Contains(t, err.Error(), "failed to write content", "error message")
				assert.Empty(t, result, "TryWrite() result on error")
			} else {
				require.NoError(t, err, "TryWrite()")
				assert.Equal(t, test.expectedResult, result, "TryWrite() result")

				// Check writer content based on type
				switch w := writer.(type) {
				case *bytes.Buffer:
					assert.Equal(t, test.expectedContent, w.String(), "buffer content")
				case *strings.Builder:
					assert.Equal(t, test.expectedContent, w.String(), "string builder content")
				}
			}
		})
	}
}

// failingWriter always returns an error on Write.
type failingWriter struct{}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestTryWriteFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		setupTest          func(t *testing.T) (content, outputPath string, force bool)
		expectError        bool
		expectedErrMessage string
		verifyResult       func(t *testing.T, tempDir, outputPath, content, result string)
	}{
		{
			name: "empty output",
			setupTest: func(_ *testing.T) (string, string, bool) {
				return testContent, "", false
			},
			expectError:        true,
			expectedErrMessage: "",
		},
		{
			name: "new file",
			setupTest: func(t *testing.T) (string, string, bool) {
				t.Helper()
				content := "new file content"
				tempDir := t.TempDir()
				outputPath := filepath.Join(tempDir, "test.txt")

				return content, outputPath, false
			},
			expectError: false,
			verifyResult: func(t *testing.T, tempDir, outputPath, content, _ string) {
				// Verify file was written
				writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
				require.NoError(t, err, "ReadFile()")
				assert.Equal(t, content, string(writtenContent), "written file content")
			},
		},
		{
			name: "existing file no force",
			setupTest: func(t *testing.T) (string, string, bool) {
				newContent := "new content"
				tempDir := t.TempDir()
				outputPath := filepath.Join(tempDir, "existing.txt")

				// Create existing file
				err := os.WriteFile(outputPath, []byte(originalContent), 0o600)
				require.NoError(t, err, "WriteFile() setup")

				return newContent, outputPath, false
			},
			expectError: false,
			verifyResult: func(t *testing.T, tempDir, outputPath, _, _ string) {
				// Verify file was NOT overwritten
				writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
				require.NoError(t, err, "ReadFile()")
				assert.Equal(
					t,
					originalContent,
					string(writtenContent),
					"file content (should not be overwritten)",
				)
			},
		},
		{
			name: "existing file force",
			setupTest: func(t *testing.T) (string, string, bool) {
				newContent := "new content forced"
				tempDir := t.TempDir()
				outputPath := filepath.Join(tempDir, "existing-force.txt")

				// Create existing file
				err := os.WriteFile(outputPath, []byte(originalContent), 0o600)
				require.NoError(t, err, "WriteFile() setup")

				return newContent, outputPath, true
			},
			expectError: false,
			verifyResult: func(t *testing.T, tempDir, outputPath, content, _ string) {
				// Verify file was overwritten
				writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
				require.NoError(t, err, "ReadFile()")
				assert.Equal(
					t,
					content,
					string(writtenContent),
					"file content (should be overwritten)",
				)
			},
		},
		{
			name: "stat error",
			setupTest: func(t *testing.T) (string, string, bool) {
				content := "content for stat error test"
				tempDir := t.TempDir()
				outputPath := filepath.Join(tempDir, "restricted", "file.txt")

				// Create a directory with no permissions to simulate stat error
				restrictedDir := filepath.Join(tempDir, "restricted")
				err := os.Mkdir(restrictedDir, 0o000)
				require.NoError(t, err, "Mkdir() setup")

				return content, outputPath, false
			},
			expectError:        true,
			expectedErrMessage: "failed to check file",
		},
		{
			name: "write error",
			setupTest: func(_ *testing.T) (string, string, bool) {
				content := "content for write error test"
				// Use a path that cannot be written to (directory that doesn't exist)
				invalidPath := "/invalid/nonexistent/deeply/nested/path/file.txt"
				return content, invalidPath, false
			},
			expectError:        true,
			expectedErrMessage: "failed to create directory",
		},
		{
			name: "file write error",
			setupTest: func(t *testing.T) (string, string, bool) {
				content := "content for file write error test"
				tempDir := t.TempDir()
				outputPath := filepath.Join(tempDir, "readonly.txt")

				// Create a directory that exists but make it read-only to cause file write failure
				err := os.WriteFile(outputPath, []byte("existing"), 0o000) // No write permissions
				require.NoError(t, err, "WriteFile() setup")

				return content, outputPath, true // force=true to skip stat check
			},
			expectError:        true,
			expectedErrMessage: "failed to write file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			content, outputPath, force := test.setupTest(t)
			result, err := ioutils.TryWriteFile(content, outputPath, force)

			if test.expectError {
				require.Error(t, err, "TryWriteFile()")
				assert.Empty(t, result, "TryWriteFile() result on error")

				if test.expectedErrMessage != "" {
					testutils.AssertErrContains(
						t,
						err,
						test.expectedErrMessage,
						"TryWriteFile() error message",
					)
				}
			} else {
				require.NoError(t, err, "TryWriteFile()")
				assert.Equal(t, content, result, "TryWriteFile()")

				if test.verifyResult != nil {
					tempDir := filepath.Dir(outputPath)
					test.verifyResult(t, tempDir, outputPath, content, result)
				}
			}
		})
	}
}

func TestGetWriterQuiet(t *testing.T) {
	t.Parallel()

	writer := ioutils.GetWriter(true)

	if writer != io.Discard {
		t.Errorf("expected io.Discard for quiet=true, got %T", writer)
	}
}

func TestGetWriterNotQuiet(t *testing.T) {
	t.Parallel()

	writer := ioutils.GetWriter(false)

	if writer != os.Stdout {
		t.Errorf("expected os.Stdout for quiet=false, got %T", writer)
	}
}

func TestWriteFileSafeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupTest     func(t *testing.T) (content, basePath, filePath string, force bool)
		expectError   bool
		expectedError error
	}{
		{
			name: "empty base path",
			setupTest: func(_ *testing.T) (string, string, string, bool) {
				return testContent, "", "/some/path/file.txt", false
			},
			expectError:   true,
			expectedError: ioutils.ErrBasePath,
		},
		{
			name: "empty file path",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				basePath := t.TempDir()
				return testContent, basePath, "", false
			},
			expectError:   true,
			expectedError: ioutils.ErrEmptyOutputPath,
		},
		{
			name: "path outside base",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				basePath := t.TempDir()
				outsidePath := "/tmp/outside.txt" // Path clearly outside basePath
				return testContent, basePath, outsidePath, false
			},
			expectError:   true,
			expectedError: ioutils.ErrPathOutsideBase,
		},
		{
			name: "new file",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				content := "new file content via WriteFileSafe"
				basePath := t.TempDir()
				filePath := filepath.Join(basePath, "newfile.txt")
				return content, basePath, filePath, false
			},
			expectError: false,
			verifyResult: func(t *testing.T, basePath, filePath, content string) {
				// Verify file was written
				writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
				require.NoError(t, err, "ReadFile")
				assert.Equal(t, content, string(writtenContent), "file content")
			},
		},
		{
			name: "existing file no force",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				newContent := "new content"
				basePath := t.TempDir()
				filePath := filepath.Join(basePath, "existing.txt")

				// Create existing file
				err := os.WriteFile(filePath, []byte(originalContent), 0o600)
				require.NoError(t, err, "WriteFile setup")

				return newContent, basePath, filePath, false
			},
			expectError: false,
			verifyResult: func(t *testing.T, basePath, filePath, _ string) {
				// Verify file was NOT overwritten
				writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
				require.NoError(t, err, "ReadFile")
				assert.Equal(
					t,
					originalContent,
					string(writtenContent),
					"file content should not be overwritten",
				)
			},
		},
		{
			name: "existing file force",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				newContent := "new content forced"
				basePath := t.TempDir()
				filePath := filepath.Join(basePath, "existing-force.txt")

				// Create existing file
				err := os.WriteFile(filePath, []byte(originalContent), 0o600)
				require.NoError(t, err, "WriteFile setup")

				return newContent, basePath, filePath, true
			},
			expectError: false,
			verifyResult: func(t *testing.T, basePath, filePath, content string) {
				// Verify file was overwritten
				writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
				require.NoError(t, err, "ReadFile")
				assert.Equal(
					t,
					content,
					string(writtenContent),
					"file content should be overwritten",
				)
			},
		},
		{
			name: "stat error",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				content := "content for stat error test"
				basePath := t.TempDir()
				filePath := filepath.Join(basePath, "restricted", "file.txt")

				// Create a directory with no permissions to simulate stat error
				restrictedDir := filepath.Join(basePath, "restricted")
				err := os.Mkdir(restrictedDir, 0o000)
				require.NoError(t, err, "Mkdir setup")

				return content, basePath, filePath, false
			},
			expectError:        true,
			expectedErrMessage: "failed to check file",
		},
		{
			name: "directory creation error",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				content := "content for directory creation error test"
				basePath := t.TempDir()
				filePath := filepath.Join(basePath, "subdir", "file.txt")

				// Create a file with the same name as the directory we want to create
				subdirPath := filepath.Join(basePath, "subdir")
				err := os.WriteFile(subdirPath, []byte("blocking file"), 0o600)
				require.NoError(t, err, "WriteFile setup to block directory creation")

				return content, basePath, filePath, true // Use force=true to skip stat check
			},
			expectError:        true,
			expectedErrMessage: "failed to create directory",
		},
		{
			name: "file write error",
			setupTest: func(t *testing.T) (string, string, string, bool) {
				content := "content for file write error test"
				basePath := t.TempDir()
				filePath := filepath.Join(basePath, "readonly.txt")

				// Create a file with no write permissions to cause write failure
				err := os.WriteFile(filePath, []byte("existing"), 0o000) // No write permissions
				require.NoError(t, err, "WriteFile setup")

				return content, basePath, filePath, true // force=true to skip stat check
			},
			expectError:        true,
			expectedErrMessage: "failed to write file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			content, basePath, filePath, force := test.setupTest(t)
			err := ioutils.WriteFileSafe(content, basePath, filePath, force)

			if test.expectError {
				require.Error(t, err, "WriteFileSafe")

				if test.expectedError != nil {
					testutils.AssertErrWrappedContains(
						t,
						err,
						test.expectedError,
						"",
						"WriteFileSafe",
					)
				}

				if test.expectedErrMessage != "" {
					testutils.AssertErrContains(t, err, test.expectedErrMessage, "WriteFileSafe")
				}
			} else {
				require.NoError(t, err, "WriteFileSafe")

				if test.verifyResult != nil {
					test.verifyResult(t, basePath, filePath, content)
				}
			}
		})
	}
}
