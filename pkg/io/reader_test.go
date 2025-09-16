package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests are intentionally minimal and explicit to keep coverage high and behavior clear.
func TestReadFileSafe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (base, filePath string)
		expectError    bool
		expectedError  error
		expectedResult string
	}{
		{
			name: "normal read",
			setup: func(t *testing.T) (string, string) {
				base := t.TempDir()
				filePath := filepath.Join(base, "file.txt")
				err := os.WriteFile(filePath, []byte("hello safe"), 0o600)
				require.NoError(t, err, "WriteFile setup")
				return base, filePath
			},
			expectError:    false,
			expectedResult: "hello safe",
		},
		{
			name: "outside base",
			setup: func(t *testing.T) (string, string) {
				base := t.TempDir()
				outside := filepath.Join(os.TempDir(), "outside-test-file.txt")
				err := os.WriteFile(outside, []byte("nope"), 0o600)
				require.NoError(t, err, "WriteFile setup")
				return base, outside
			},
			expectError:   true,
			expectedError: ioutils.ErrPathOutsideBase,
		},
		{
			name: "traversal attempt",
			setup: func(t *testing.T) (string, string) {
				base := t.TempDir()
				parent := filepath.Join(base, "..", "traversal.txt")
				absParent, _ := filepath.Abs(parent)
				err := os.WriteFile(absParent, []byte("traversal"), 0o600)
				require.NoError(t, err, "WriteFile setup parent")
				attempt := filepath.Join(base, "..", "traversal.txt")
				return base, attempt
			},
			expectError:   true,
			expectedError: ioutils.ErrPathOutsideBase,
		},
		{
			name: "missing file inside base",
			setup: func(t *testing.T) (string, string) {
				base := t.TempDir()
				missing := filepath.Join(base, "missing.txt")
				return base, missing
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			base, filePath := tt.setup(t)
			got, err := ioutils.ReadFileSafe(base, filePath)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					testutils.AssertErrWrappedContains(t, err, tt.expectedError, "", "ReadFileSafe")
				} else {
					testutils.AssertErrContains(t, err, "failed to read file", "ReadFileSafe")
				}
			} else {
				require.NoError(t, err, "ReadFileSafe")
				assert.Equal(t, tt.expectedResult, string(got), "content")
			}
		})
	}
}
