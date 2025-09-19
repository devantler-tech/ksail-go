package validators_test

import (
	"errors"
	"testing"

	validators "github.com/devantler-tech/ksail-go/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockValidator is a simple implementation of the Validator interface for testing.
type mockValidator struct {
	shouldError bool
	errorMsg    string
}

func (m *mockValidator) Validate() error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	return nil
}

// TestValidatorInterface tests that the Validator interface can be implemented.
func TestValidatorInterface(t *testing.T) {
	t.Parallel()

	t.Run("successful validation", func(t *testing.T) {
		t.Parallel()

		var validator validators.Validator = &mockValidator{shouldError: false}
		err := validator.Validate()

		require.NoError(t, err)
	})

	t.Run("validation with error", func(t *testing.T) {
		t.Parallel()

		expectedError := "validation failed"
		var validator validators.Validator = &mockValidator{
			shouldError: true,
			errorMsg:    expectedError,
		}
		err := validator.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedError)
	})
}

// TestValidatorInterfaceStructure tests the interface structure and method signature.
func TestValidatorInterfaceStructure(t *testing.T) {
	t.Parallel()

	// Test that mockValidator implements the Validator interface
	var _ validators.Validator = (*mockValidator)(nil)

	// Test basic functionality
	validator := &mockValidator{shouldError: false}
	err := validator.Validate()
	assert.NoError(t, err)
}
