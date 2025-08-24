package testutil_test

import (
	"sync/atomic"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutil"
)

// TestDefaultNameCases ensures the helper returns the expected two standard cases.
func TestDefaultNameCases(t *testing.T) {
	t.Parallel()

	// Arrange
	defaultName := "cfg-default"

	// Act
	cases := testutil.DefaultNameCases(defaultName)

	// Assert
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}

	if cases[0].Name == "" || cases[0].InputName == "" || cases[0].ExpectedName == "" {
		t.Fatalf("first case fields should not be empty: %#v", cases[0])
	}

	if cases[1].InputName != "" || cases[1].ExpectedName != defaultName {
		t.Fatalf("second case should use default name when input empty; got: %#v", cases[1])
	}
}

// TestRunNameCases_RunsAllCases ensures the runner executes provided function for each case.
func TestRunNameCases_RunsAllCases(t *testing.T) {
	t.Parallel()

	// Arrange
	cases := []testutil.NameCase{
		{Name: "case1", InputName: "a", ExpectedName: "a"},
		{Name: "case2", InputName: "", ExpectedName: "def"},
		{Name: "case3", InputName: "x", ExpectedName: "x"},
	}

	var ran int64
	// Assert (after subtests complete)
	t.Cleanup(func() {
		if got := atomic.LoadInt64(&ran); got != int64(len(cases)) {
			t.Fatalf("expected runner to execute %d cases, ran %d", len(cases), got)
		}
	})

	// Act
	testutil.RunNameCases(t, cases, func(t *testing.T, c testutil.NameCase) {
		t.Helper()
		// Assert (per-case)
		if c.InputName == "" && c.ExpectedName == "" {
			t.Fatalf("expected ExpectedName to be set when InputName is empty: %#v", c)
		}

		atomic.AddInt64(&ran, 1)
	})
}
