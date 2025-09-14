package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionVariables(t *testing.T) {
	t.Parallel()

	// Test that version variables are initialized with default values
	assert.Equal(t, "dev", version)
	assert.Equal(t, "none", commit)
	assert.Equal(t, "unknown", date)
}

// Note: Testing main() directly is challenging because it calls os.Exit()
// The main function is covered by integration tests through cmd.Execute()
// which is already tested in the cmd package.
