// Package clustertestutils provides common test utilities for cluster provisioner testing,
// including shared test cases and helper functions for standardizing test patterns.
package clustertestutils

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
)

// Common error variables used across cluster provisioner tests to avoid duplication
var (
	ErrCreateClusterFailed = errors.New("create cluster failed")
	ErrDeleteClusterFailed = errors.New("delete cluster failed")
	ErrListClustersFailed  = errors.New("list clusters failed")
	ErrStartClusterFailed  = errors.New("start cluster failed")
	ErrStopClusterFailed   = errors.New("stop cluster failed")
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