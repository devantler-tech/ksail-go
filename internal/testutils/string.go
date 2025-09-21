package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertStringContains asserts that the haystack contains all provided substrings.
func AssertStringContains(t *testing.T, haystack string, subs ...string) {
	t.Helper()

	for _, s := range subs {
		assert.Contains(t, haystack, s)
	}
}

// AssertStringContainsOneOf asserts that the haystack contains at least one of the provided options.
func AssertStringContainsOneOf(t *testing.T, haystack string, options ...string) {
	t.Helper()

	for _, opt := range options {
		if assert.Contains(t, haystack, opt) {
			return
		}
	}

	assert.Failf(
		t,
		"string did not contain any expected option",
		"wanted one of %v in: %s",
		options,
		haystack,
	)
}
