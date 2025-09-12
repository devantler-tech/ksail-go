package testutils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
)

func TestAssertErrWrappedContains(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	t.Run("with_valid_wrapped_error_and_substring", func(t *testing.T) {
		t.Parallel()
		// Should not panic - this tests the happy path
		testutils.AssertErrWrappedContains(t, wrappedErr, baseErr, "wrapped", "test context")
	})

	t.Run("with_valid_wrapped_error_no_substring", func(t *testing.T) {
		t.Parallel()
		// Should not panic - this tests the case with empty contains string
		testutils.AssertErrWrappedContains(t, wrappedErr, baseErr, "", "test context")
	})

	t.Run("with_nil_want_error", func(t *testing.T) {
		t.Parallel()
		// Should not panic - this tests the case with nil want error (just check for any error)
		testutils.AssertErrWrappedContains(t, wrappedErr, nil, "wrapped", "test context")
	})
}

func TestAssertErrContains(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error message")

	t.Run("with_error_and_substring", func(t *testing.T) {
		t.Parallel()
		// Should not panic - this tests the happy path
		testutils.AssertErrContains(t, testErr, "test error", "test context")
	})

	t.Run("with_error_and_empty_substring", func(t *testing.T) {
		t.Parallel()
		// Should not panic - this tests the case with empty contains string
		testutils.AssertErrContains(t, testErr, "", "test context")
	})
}
