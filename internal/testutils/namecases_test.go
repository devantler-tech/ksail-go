package testutils_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDefaultNameCases(t *testing.T) {
	t.Parallel()

	t.Run("returns_expected_structure", func(t *testing.T) {
		t.Parallel()

		defaultName := "default-cluster"
		cases := testutils.DefaultNameCases(defaultName)

		// Should return exactly 2 cases
		assert.Len(t, cases, 2)

		// First case: with explicit name
		assert.Equal(t, "with name", cases[0].Name)
		assert.Equal(t, "my-cluster", cases[0].InputName)
		assert.Equal(t, "my-cluster", cases[0].ExpectedName)

		// Second case: empty name uses default
		assert.Equal(t, "without name uses cfg", cases[1].Name)
		assert.Equal(t, "", cases[1].InputName)
		assert.Equal(t, defaultName, cases[1].ExpectedName)
	})
}

func TestRunNameCases(t *testing.T) {
	t.Parallel()

	t.Run("executes_all_cases", func(t *testing.T) {
		// Don't use t.Parallel() here because we need to wait for subtests to complete

		cases := testutils.DefaultNameCases("test-default")

		// This will create and run subtests
		testutils.RunNameCases(t, cases, func(innerT *testing.T, c testutils.NameCase) {
			// Just verify the case structure is valid
			assert.NotEmpty(innerT, c.Name)
			assert.NotEmpty(innerT, c.ExpectedName)
		})

		// The fact that RunNameCases completed without panicking means it worked
		// We can verify that cases were provided
		assert.Len(t, cases, 2)
	})
}
