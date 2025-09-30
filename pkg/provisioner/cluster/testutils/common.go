// Package clustertestutils provides common test utilities for cluster provisioner testing,
// including shared test cases and helper functions for standardizing test patterns.
package clustertestutils

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
)

// Common error variables used across cluster provisioner tests to avoid duplication.
var (
	ErrCreateClusterFailed  = errors.New("create cluster failed")
	ErrDeleteClusterFailed  = errors.New("delete cluster failed")
	ErrListClustersFailed   = errors.New("list clusters failed")
	ErrStartClusterFailed   = errors.New("start cluster failed")
	ErrStopClusterFailed    = errors.New("stop cluster failed")
	ErrScaleNodeGroupFailed = errors.New("scale node group failed")
)

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
// - Getting test cases
// - Running testutils.RunNameCases with t.Helper().
// Note: Caller should call t.Parallel() for parallel execution.
func RunStandardSuccessTest(
	t *testing.T,
	cases []testutils.NameCase,
	testRunner func(t *testing.T, inputName, expectedName string),
) {
	t.Helper()

	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		testRunner(t, nameCase.InputName, nameCase.ExpectedName)
	})
}

// RunCreateTest runs a standard Create success test pattern with the common cfg-name cases.
// This centralizes the Create test pattern of:
// - Getting DefaultNameCases("cfg-name")
// - Running RunStandardSuccessTest with "Create()" action.
func RunCreateTest(
	t *testing.T,
	runActionSuccessFunc func(t *testing.T, inputName, expectedName string),
) {
	t.Helper()

	cases := DefaultNameCases("cfg-name")
	RunStandardSuccessTest(t, cases, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		runActionSuccessFunc(t, inputName, expectedName)
	})
}

// RunActionSuccess provides a generic helper for testing successful provisioner actions.
// It eliminates code duplication between Kind and K3d test files by abstracting the common pattern:
// setup -> expect -> action -> assert.
func RunActionSuccess[MockT, ProvisionerT any](
	t *testing.T,
	label string,
	inputName, expectedName string,
	setupFn func(*testing.T) (ProvisionerT, MockT),
	expectFn func(MockT, string),
	actionFn func(ProvisionerT, string) error,
) {
	t.Helper()
	provisioner, mock := setupFn(t)
	expectFn(mock, expectedName)

	err := actionFn(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

// RunCreateSuccessTest provides a standard test pattern for Create operations.
// This eliminates duplication between Kind and K3d test files by providing the common
// TestCreate_Success structure that both can use.
// Note: The caller should call t.Parallel() for parallel execution.
func RunCreateSuccessTest[MockT, ProvisionerT any](
	t *testing.T,
	setupFn func(*testing.T) (ProvisionerT, MockT),
	expectFn func(MockT, string),
	actionFn func(ProvisionerT, string) error,
) {
	t.Helper()
	RunCreateTest(t, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		RunActionSuccess(
			t,
			"Create()",
			inputName,
			expectedName,
			setupFn,
			expectFn,
			actionFn,
		)
	})
}
