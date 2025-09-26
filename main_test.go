package main

import (
	"os"
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

func TestRunSuccess(t *testing.T) {
	t.Parallel()

	// Test run function with --help flag (this typically exits with 0)
	oldArgs := os.Args

	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail", "--help"}

	exitCode := run()

	// --help should exit with code 0 (even though it prints usage and exits)
	assert.Equal(t, 0, exitCode)
}

func TestRunWithInvalidCommand(t *testing.T) {
	t.Parallel()

	// Test run function with invalid command
	oldArgs := os.Args

	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail", "invalid-command"}

	exitCode := run()

	// In test environment, invalid command shows help and exits with 0
	assert.Equal(t, 0, exitCode)
}
