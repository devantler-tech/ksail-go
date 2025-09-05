package clustertestutils

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
)

// ErrBoom is a common test error used across cluster provisioner tests.
var ErrBoom = errors.New("boom")

// DefaultDeleteCases returns standard test cases for testing delete operations with name handling.
func DefaultDeleteCases() []testutils.NameCase {
	return []testutils.NameCase{
		{Name: "without name uses cfg", InputName: "", ExpectedName: "cfg-name"},
		{Name: "with name", InputName: "custom", ExpectedName: "custom"},
	}
}

// DefaultNameCases returns standard test cases for testing operations with name handling.
func DefaultNameCases(cfgName string) []testutils.NameCase {
	return []testutils.NameCase{
		{Name: "without name uses cfg", InputName: "", ExpectedName: cfgName},
		{Name: "with name", InputName: "custom", ExpectedName: "custom"},
	}
}

// RunStandardSuccessTest runs a standard success test pattern with parallel execution and name cases.
// This centralizes the common pattern of:
// - t.Parallel()
// - Getting test cases
// - Running testutils.RunNameCases with t.Helper()
func RunStandardSuccessTest(
	t *testing.T,
	cases []testutils.NameCase,
	testRunner func(t *testing.T, inputName, expectedName string),
) {
	t.Helper()
	t.Parallel()

	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		testRunner(t, nameCase.InputName, nameCase.ExpectedName)
	})
}