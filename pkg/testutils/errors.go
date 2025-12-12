package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Error assertion helpers.

// AssertErrWrappedContains verifies that an error exists, wraps a target error,
// and optionally contains a given substring in its message.
// The ctx parameter describes the calling context for better error messages.
func AssertErrWrappedContains(t *testing.T, got error, want error, contains string, ctx string) {
	t.Helper()

	// Assert the error type/unwrap match first
	if want != nil {
		require.ErrorIs(t, got, want, ctx)
	} else {
		require.Error(t, got, ctx)
	}

	// And optionally assert the message content
	if contains != "" {
		assert.ErrorContains(t, got, contains, ctx)
	}
}

// AssertErrContains asserts that an error is non-nil and its message contains the provided substring.
// If contains is empty, it only checks that the error is non-nil.
func AssertErrContains(t *testing.T, got error, contains string, ctx string) {
	t.Helper()

	if contains == "" {
		assert.Error(t, got, ctx)

		return
	}

	// ErrorContains also asserts non-nil
	assert.ErrorContains(t, got, contains, ctx)
}
