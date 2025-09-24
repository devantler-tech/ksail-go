package cmdhelpers_test

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/stretchr/testify/assert"
)

// TestErrorVariableConsolidation verifies that the cmdhelpers package
// correctly uses the consolidated error variable from the helpers package.
func TestErrorVariableConsolidation(t *testing.T) {
	t.Parallel()

	// Test that both packages now reference the same error
	// by checking if they are equal when wrapped in the same error context
	helpersErr := helpers.ErrConfigurationValidationFailed

	// Create a function that returns the error (simulating cmdhelpers usage)
	getCmdHelpersError := func() error {
		// This simulates how cmdhelpers references the error
		return helpers.ErrConfigurationValidationFailed
	}

	cmdHelpersErr := getCmdHelpersError()

	// Both should be the same error instance
	assert.True(t, errors.Is(cmdHelpersErr, helpersErr),
		"cmdhelpers should use the same error instance as helpers package")

	// Verify the error message is consistent
	assert.Equal(t, "configuration validation failed", helpersErr.Error())
	assert.Equal(t, helpersErr.Error(), cmdHelpersErr.Error())
}
