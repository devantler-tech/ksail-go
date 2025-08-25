// Package testutils provides testing utilities to aid with assertions in tests.
package testutils

import (
	"reflect"
	"testing"
)

// AssertNoError fails the test if err is not nil with a contextual label.
func AssertNoError(t *testing.T, err error, ctx string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s unexpected error: %v", ctx, err)
	}
}

// AssertStringsEqualOrder asserts that two string slices are equal (same order).
func AssertStringsEqualOrder(t *testing.T, got, want []string, ctx string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s got %v, want %v", ctx, got, want)
	}
}
