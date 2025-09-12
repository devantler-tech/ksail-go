package testutils_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
)

func TestAssertStringContains(t *testing.T) {
	t.Parallel()

	t.Run("all_substrings_present", func(t *testing.T) {
		t.Parallel()

		haystack := "this is a test string with multiple words"

		// Should not panic - all substrings are present
		testutils.AssertStringContains(t, haystack, "test", "string", "multiple")
	})

	t.Run("single_substring", func(t *testing.T) {
		t.Parallel()

		haystack := "hello world"

		// Should not panic - single substring is present
		testutils.AssertStringContains(t, haystack, "world")
	})

	t.Run("no_substrings", func(t *testing.T) {
		t.Parallel()

		haystack := "hello world"

		// Should not panic - no substrings to check
		testutils.AssertStringContains(t, haystack)
	})
}

func TestAssertStringContainsOneOf(t *testing.T) {
	t.Parallel()

	t.Run("first_option_matches", func(t *testing.T) {
		t.Parallel()

		haystack := "hello world"

		// Should not panic - first option matches (no error logs)
		testutils.AssertStringContainsOneOf(t, haystack, "hello", "goodbye", "test")
	})

	t.Run("one_option_matches", func(t *testing.T) {
		t.Parallel()

		haystack := "hello world"

		// Should not panic - at least one option matches
		testutils.AssertStringContainsOneOf(t, haystack, "world")
	})
}
