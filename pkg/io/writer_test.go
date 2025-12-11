package io_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
)

const (
	testContent     = "test content"
	originalContent = "original content"
)

func TestTryWrite(t *testing.T) {
	t.Parallel()

	tests := getTryWriteTests()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			runTryWriteTest(t, test)
		})
	}
}

func getTryWriteTests() []struct {
	name            string
	content         string
	setupWriter     func() io.Writer
	expectError     bool
	expectedResult  string
	expectedContent string
} {
	return []struct {
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
}

func runTryWriteTest(t *testing.T, test struct {
	name            string
	content         string
	setupWriter     func() io.Writer
	expectError     bool
	expectedResult  string
	expectedContent string
},
) {
	t.Helper()

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
}

// failingWriter always returns an error on Write.
type failingWriter struct{}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestTryWriteFile(t *testing.T) {
	t.Parallel()

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()
		runTryWriteFileValidationTests(t)
	})

	t.Run("successful operations", func(t *testing.T) {
		t.Parallel()
		runTryWriteFileSuccessTests(t)
	})

	t.Run("file system errors", func(t *testing.T) {
		t.Parallel()
		runTryWriteFileErrorTests(t)
	})
}

func runTryWriteFileValidationTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		content    string
		outputPath string
		force      bool
	}{
		{
			name:       "empty output",
			content:    testContent,
			outputPath: "",
			force:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := ioutils.TryWriteFile(test.content, test.outputPath, test.force)

			require.Error(t, err, "TryWriteFile()")
			assert.Empty(t, result, "TryWriteFile() result on error")
		})
	}
}

func runTryWriteFileSuccessTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name         string
		setupTest    func(t *testing.T) (content, outputPath string, force bool)
		verifyResult func(t *testing.T, tempDir, outputPath, content, result string)
	}{
		{
			name:         "new file",
			setupTest:    setupTryWriteFileNewFile,
			verifyResult: verifyTryWriteFileContentsEqual,
		},
		{
			name:         "existing file no force",
			setupTest:    setupTryWriteFileExistingNoForce,
			verifyResult: verifyTryWriteFileContentsPreserved,
		},
		{
			name:         "existing file force",
			setupTest:    setupTryWriteFileExistingForce,
			verifyResult: verifyTryWriteFileContentsEqual,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			content, outputPath, force := test.setupTest(t)
			result, err := ioutils.TryWriteFile(content, outputPath, force)

			require.NoError(t, err, "TryWriteFile()")
			assert.Equal(t, content, result, "TryWriteFile()")

			if test.verifyResult != nil {
				tempDir := filepath.Dir(outputPath)
				test.verifyResult(t, tempDir, outputPath, content, result)
			}
		})
	}
}

func setupTryWriteFileNewFile(t *testing.T) (string, string, bool) {
	t.Helper()

	content := "new file content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.txt")

	return content, outputPath, false
}

func setupTryWriteFileExistingNoForce(t *testing.T) (string, string, bool) {
	t.Helper()

	newContent := "new content"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.txt")

	err := os.WriteFile(outputPath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile() setup")

	return newContent, outputPath, false
}

func setupTryWriteFileExistingForce(t *testing.T) (string, string, bool) {
	t.Helper()

	newContent := "new content forced"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing-force.txt")

	err := os.WriteFile(outputPath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile() setup")

	return newContent, outputPath, true
}

func verifyTryWriteFileContentsEqual(t *testing.T, tempDir, outputPath, content, _ string) {
	t.Helper()

	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")
	assert.Equal(t, content, string(writtenContent), "written file content")
}

func verifyTryWriteFileContentsPreserved(t *testing.T, tempDir, outputPath, _, _ string) {
	t.Helper()

	writtenContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "ReadFile()")

	assert.Equal(t, originalContent, string(writtenContent),
		"file content (should not be overwritten)")
}

func runTryWriteFileErrorTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name               string
		setupTest          func(t *testing.T) (content, outputPath string, force bool)
		expectedErrMessage string
	}{
		{
			name:               "stat error",
			setupTest:          setupTryWriteFileStatError,
			expectedErrMessage: "failed to check file",
		},
		{
			name:               "write error",
			setupTest:          setupTryWriteFileWriteError,
			expectedErrMessage: "failed to create directory",
		},
		{
			name:               "file write error",
			setupTest:          setupTryWriteFilePermissionError,
			expectedErrMessage: "failed to write file",
		},
	}

	runErrorTestsWithTwoParams(t, tests, ioutils.TryWriteFile, "TryWriteFile")
}

func setupTryWriteFileStatError(t *testing.T) (string, string, bool) {
	t.Helper()

	content := "content for stat error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "restricted", "file.txt")

	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0o000)
	require.NoError(t, err, "Mkdir() setup")

	return content, outputPath, false
}

func setupTryWriteFileWriteError(_ *testing.T) (string, string, bool) {
	content := "content for write error test"
	invalidPath := "/invalid/nonexistent/deeply/nested/path/file.txt"

	return content, invalidPath, false
}

func setupTryWriteFilePermissionError(t *testing.T) (string, string, bool) {
	t.Helper()

	content := "content for file write error test"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "readonly.txt")

	err := os.WriteFile(outputPath, []byte("existing"), 0o000)
	require.NoError(t, err, "WriteFile() setup")

	return content, outputPath, true
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

func TestWriteFileSafe(t *testing.T) {
	t.Parallel()

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()
		runWriteFileSafeValidationTests(t)
	})

	t.Run("successful operations", func(t *testing.T) {
		t.Parallel()
		runWriteFileSafeSuccessTests(t)
	})

	t.Run("file system errors", func(t *testing.T) {
		t.Parallel()
		runWriteFileSafeFileSystemErrorTests(t)
	})
}

func runWriteFileSafeValidationTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name          string
		content       string
		basePath      string
		filePath      string
		force         bool
		expectedError error
	}{
		{
			name:          "empty base path",
			content:       testContent,
			basePath:      "",
			filePath:      "/some/path/file.txt",
			force:         false,
			expectedError: ioutils.ErrBasePath,
		},
		{
			name:          "empty file path",
			content:       testContent,
			basePath:      "/tmp",
			filePath:      "",
			force:         false,
			expectedError: ioutils.ErrEmptyOutputPath,
		},
		{
			name:          "path outside base",
			content:       testContent,
			basePath:      "/tmp/test",
			filePath:      "/tmp/outside.txt",
			force:         false,
			expectedError: ioutils.ErrPathOutsideBase,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := ioutils.WriteFileSafe(test.content, test.basePath, test.filePath, test.force)

			require.Error(t, err, "WriteFileSafe")
			testutils.AssertErrWrappedContains(t, err, test.expectedError, "", "WriteFileSafe")
		})
	}
}

func runWriteFileSafeSuccessTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name         string
		setupTest    func(t *testing.T) (content, basePath, filePath string, force bool)
		verifyResult func(t *testing.T, basePath, filePath, content string)
	}{
		{
			name:         "new file",
			setupTest:    setupWriteFileSafeNewFile,
			verifyResult: verifyFileContentsEqual,
		},
		{
			name:         "existing file no force",
			setupTest:    setupWriteFileSafeExistingNoForce,
			verifyResult: verifyFileContentsPreserved,
		},
		{
			name:         "existing file force",
			setupTest:    setupWriteFileSafeExistingForce,
			verifyResult: verifyFileContentsEqual,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			content, basePath, filePath, force := test.setupTest(t)
			err := ioutils.WriteFileSafe(content, basePath, filePath, force)

			require.NoError(t, err, "WriteFileSafe")

			if test.verifyResult != nil {
				test.verifyResult(t, basePath, filePath, content)
			}
		})
	}
}

func setupWriteFileSafeNewFile(t *testing.T) (string, string, string, bool) {
	t.Helper()

	content := "new file content via WriteFileSafe"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "newfile.txt")

	return content, basePath, filePath, false
}

func setupWriteFileSafeExistingNoForce(t *testing.T) (string, string, string, bool) {
	t.Helper()

	newContent := "new content"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "existing.txt")

	err := os.WriteFile(filePath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile setup")

	return newContent, basePath, filePath, false
}

func setupWriteFileSafeExistingForce(t *testing.T) (string, string, string, bool) {
	t.Helper()

	newContent := "new content forced"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "existing-force.txt")

	err := os.WriteFile(filePath, []byte(originalContent), 0o600)
	require.NoError(t, err, "WriteFile setup")

	return newContent, basePath, filePath, true
}

func verifyFileContentsEqual(t *testing.T, basePath, filePath, content string) {
	t.Helper()

	writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
	require.NoError(t, err, "ReadFile")
	assert.Equal(t, content, string(writtenContent), "file content")
}

func verifyFileContentsPreserved(t *testing.T, basePath, filePath, _ string) {
	t.Helper()

	writtenContent, err := ioutils.ReadFileSafe(basePath, filePath)
	require.NoError(t, err, "ReadFile")

	assert.Equal(t, originalContent, string(writtenContent),
		"file content should not be overwritten")
}

func runWriteFileSafeFileSystemErrorTests(t *testing.T) {
	t.Helper()

	tests := []struct {
		name               string
		setupTest          func(t *testing.T) (content, basePath, filePath string, force bool)
		expectedErrMessage string
	}{
		{
			name:               "stat error",
			setupTest:          setupStatErrorTest,
			expectedErrMessage: "failed to check file",
		},
		{
			name:               "directory creation error",
			setupTest:          setupDirectoryCreationErrorTest,
			expectedErrMessage: "failed to create directory",
		},
		{
			name:               "file write error",
			setupTest:          setupFileWriteErrorTest,
			expectedErrMessage: "failed to write file",
		},
	}

	runErrorTestsWithThreeParams(t, tests, ioutils.WriteFileSafe, "WriteFileSafe")
}

func setupStatErrorTest(t *testing.T) (string, string, string, bool) {
	t.Helper()

	content := "content for stat error test"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "restricted", "file.txt")

	restrictedDir := filepath.Join(basePath, "restricted")
	err := os.Mkdir(restrictedDir, 0o000)
	require.NoError(t, err, "Mkdir setup")

	return content, basePath, filePath, false
}

func setupDirectoryCreationErrorTest(t *testing.T) (string, string, string, bool) {
	t.Helper()

	content := "content for directory creation error test"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "subdir", "file.txt")

	subdirPath := filepath.Join(basePath, "subdir")
	err := os.WriteFile(subdirPath, []byte("blocking file"), 0o600)
	require.NoError(t, err, "WriteFile setup to block directory creation")

	return content, basePath, filePath, true
}

func setupFileWriteErrorTest(t *testing.T) (string, string, string, bool) {
	t.Helper()

	content := "content for file write error test"
	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "readonly.txt")

	err := os.WriteFile(filePath, []byte("existing"), 0o000)
	require.NoError(t, err, "WriteFile setup")

	return content, basePath, filePath, true
}

// runErrorTestsWithTwoParams runs error tests for functions with two parameters (content, outputPath).
func runErrorTestsWithTwoParams(
	t *testing.T,
	tests []struct {
		name               string
		setupTest          func(t *testing.T) (content, outputPath string, force bool)
		expectedErrMessage string
	},
	testFunc func(content, outputPath string, force bool) (string, error),
	functionName string,
) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			content, outputPath, force := test.setupTest(t)
			result, err := testFunc(content, outputPath, force)

			require.Error(t, err, functionName)
			assert.Empty(t, result, functionName+" result on error")
			testutils.AssertErrContains(t, err, test.expectedErrMessage, functionName)
		})
	}
}

// runErrorTestsWithThreeParams runs error tests for functions with three parameters (content, basePath, filePath).
func runErrorTestsWithThreeParams(
	t *testing.T,
	tests []struct {
		name               string
		setupTest          func(t *testing.T) (content, basePath, filePath string, force bool)
		expectedErrMessage string
	},
	testFunc func(content, basePath, filePath string, force bool) error,
	functionName string,
) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			content, basePath, filePath, force := test.setupTest(t)
			err := testFunc(content, basePath, filePath, force)

			require.Error(t, err, functionName)
			testutils.AssertErrContains(t, err, test.expectedErrMessage, functionName)
		})
	}
}
