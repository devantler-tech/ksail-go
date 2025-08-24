package testutil_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutil"
)

// TestAssertNoError_Succeeds ensures AssertNoError does not fail when err is nil.
func TestAssertNoError_Succeeds(t *testing.T) {
	t.Parallel()
	// Arrange
	var err error

	// Act
	testutil.AssertNoError(t, err, "no error expected")
	// Assert
	// If AssertNoError called t.Fatalf, this test would fail; reaching here means success.
}

// TestAssertStringsEqualOrder_Succeeds ensures AssertStringsEqualOrder passes for equal slices.
func TestAssertStringsEqualOrder_Succeeds(t *testing.T) {
	t.Parallel()
	// Arrange
	got := []string{"a", "b", "c"}
	want := []string{"a", "b", "c"}

	// Act
	testutil.AssertStringsEqualOrder(t, got, want, "equal slices")
	// Assert
	// If AssertStringsEqualOrder called t.Fatalf, this test would fail; reaching here means success.
}
