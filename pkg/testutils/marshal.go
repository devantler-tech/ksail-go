package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
)

// MustMarshal marshals a value or fails the test.
func MustMarshal[T any](t *testing.T, m marshaller.Marshaller[T], v T) string {
	t.Helper()

	s, err := m.Marshal(v)
	require.NoError(t, err, "MustMarshal should not error")

	return s
}

// MustUnmarshal unmarshals bytes into a value or fails the test.
func MustUnmarshal[T any](t *testing.T, m marshaller.Marshaller[T], data []byte, out *T) {
	t.Helper()

	err := m.Unmarshal(data, out)
	require.NoError(t, err, "MustUnmarshal should not error")
}

// MustUnmarshalString unmarshals a string into a value or fails the test.
func MustUnmarshalString[T any](t *testing.T, m marshaller.Marshaller[T], data string, out *T) {
	t.Helper()

	err := m.UnmarshalString(data, out)
	require.NoError(t, err, "MustUnmarshalString should not error")
}
