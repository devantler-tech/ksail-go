package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
)

// AssertFileEquals compares the file content with the expected string.
func AssertFileEquals(t *testing.T, dir, path, expected string) {
	t.Helper()

	fileContent, err := ioutils.ReadFileSafe(dir, path)

	require.NoError(t, err, "File should exist")
	assert.Equal(t, expected, string(fileContent))
}
