// Package testutil provides testing utilities to aid error handling in tests.
package testutil

import (
	"errors"
	"strings"
	"testing"
)

// AssertErrWrappedContains verifies that an error exists, wraps a target error,
// and optionally contains a given substring in its message. "ctx" describes the calling context.
func AssertErrWrappedContains(t *testing.T, got error, want error, contains string, ctx string) {
    t.Helper()

    if got == nil {
        t.Fatalf("%s expected error, got nil", ctx)
    }

    if !errors.Is(got, want) {
        t.Fatalf("%s error = %v, want wrapped %v", ctx, got, want)
    }

    if contains != "" && !strings.Contains(got.Error(), contains) {
        t.Fatalf("%s error message = %q, want to contain %s", ctx, got.Error(), contains)
    }
}
