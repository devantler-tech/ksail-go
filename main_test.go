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

func TestRunBasic(t *testing.T) {
	t.Parallel()

	// Test that run function works without panicking with no arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail"}

	assert.NotPanics(t, func() {
		run()
	}, "run() should not panic")
}

func TestRunWithHelp(t *testing.T) {
	t.Parallel()

	// Test that run function handles help flag without panicking
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail", "--help"}

	assert.NotPanics(t, func() {
		run()
	}, "run() with help should not panic")
}

func TestRunWithInvalidCommand(t *testing.T) {
	t.Parallel()

	// Test that run function handles invalid commands without panicking
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail", "invalid-command"}

	assert.NotPanics(t, func() {
		run()
	}, "run() with invalid command should not panic")
}

func TestRunWithVersionFlag(t *testing.T) {
	t.Parallel()

	// Test that run function handles version flag without panicking
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail", "--version"}

	assert.NotPanics(t, func() {
		run()
	}, "run() with version should not panic")
}

func TestRunWithSubcommandHelp(t *testing.T) {
	t.Parallel()

	// Test that run function handles init help without panicking
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ksail", "init", "--help"}

	assert.NotPanics(t, func() {
		run()
	}, "run() with init help should not panic")
}

func TestMainFunction(t *testing.T) {
	// Test that main() doesn't panic when called
	// Note: We can't easily test main() directly as it may call os.Exit()
	// but we can verify it doesn't panic when run is called properly
	assert.NotPanics(t, func() {
		// Simulate what main does
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"ksail", "--help"}
		run()
	}, "main() simulation should not panic")
}
