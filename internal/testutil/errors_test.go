package testutil_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutil"
)

var errBaseFailure = errors.New("base failure")
var errTarget = errors.New("target error")

// TestAssertErrWrappedContains_WithContains validates behavior when substring match is required.
func TestAssertErrWrappedContains_WithContains(t *testing.T) {
	t.Parallel()

	// Arrange
	base := errBaseFailure
	wrapped := fmt.Errorf("operation failed: %w", base)

	// Act
	testutil.AssertErrWrappedContains(t, wrapped, base, "operation failed", "wrapped contains")
	// Assert
	// If the helper fails, this test will fail. Reaching here means success.
}

// TestAssertErrWrappedContains_WithoutContains validates behavior when no substring check requested.
func TestAssertErrWrappedContains_WithoutContains(t *testing.T) {
	t.Parallel()

	// Arrange
	target := errTarget
	got := fmt.Errorf("another context: %w", target)

	// Act
	testutil.AssertErrWrappedContains(t, got, target, "", "wrapped no contains")
	// Assert
	// Success if no fatal failure.
}
