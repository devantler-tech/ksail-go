package helpers_test

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// Common test cases used by multiple formatter tests to eliminate duplication.

// TestCase represents a test case for formatting functions.
type TestCase struct {
	Name     string
	Result   *validator.ValidationResult
	Expected string
}
