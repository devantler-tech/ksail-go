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

	t.Run("normal read", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		filePath := filepath.Join(base, "file.txt")
		want := "hello safe"
		err := os.WriteFile(filePath, []byte(want), 0o600)
		require.NoError(t, err, "WriteFile setup")

		got, err := ioutils.ReadFileSafe(base, filePath)

		require.NoError(t, err, "ReadFileSafe")
		assert.Equal(t, want, string(got), "content")
	})

	t.Run("outside base", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		outside := filepath.Join(os.TempDir(), "outside-test-file.txt")
		err := os.WriteFile(outside, []byte("nope"), 0o600)
		require.NoError(t, err, "WriteFile setup")

		_, err = ioutils.ReadFileSafe(base, outside)

		testutils.AssertErrWrappedContains(t, err, ioutils.ErrPathOutsideBase, "", "ReadFileSafe")
	})

	t.Run("traversal attempt", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		parent := filepath.Join(base, "..", "traversal.txt")
		absParent, _ := filepath.Abs(parent)
		err := os.WriteFile(absParent, []byte("traversal"), 0o600)
		require.NoError(t, err, "WriteFile setup parent")

		attempt := filepath.Join(base, "..", "traversal.txt")

		_, err = ioutils.ReadFileSafe(base, attempt)

		testutils.AssertErrWrappedContains(t, err, ioutils.ErrPathOutsideBase, "", "ReadFileSafe")
	})

	t.Run("missing file inside base", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		missing := filepath.Join(base, "missing.txt")

		_, err := ioutils.ReadFileSafe(base, missing)

		testutils.AssertErrContains(t, err, "failed to read file", "ReadFileSafe")
	})
}
