package testutils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ErrTestConfigLoadError is a static test error to comply with err113 and allow reuse across packages.
var ErrTestConfigLoadError = errors.New("test config load error")

// AssertErrWrappedContains verifies that an error exists, wraps a target error,
// and optionally contains a given substring in its message. "ctx" describes the calling context.
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
func AssertErrContains(t *testing.T, got error, contains string, ctx string) {
	t.Helper()

	if contains == "" {
		assert.Error(t, got, ctx)

		return
	}

	// ErrorContains also asserts non-nil
	assert.ErrorContains(t, got, contains, ctx)
}
