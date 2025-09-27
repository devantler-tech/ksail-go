package main

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:gochecknoglobals // Shared across tests to synchronize os.Args mutations.
var osArgsMu sync.Mutex

func withArgs(t *testing.T, args []string, runner func()) {
	t.Helper()

	osArgsMu.Lock()

	oldArgs := os.Args
	os.Args = args

	defer func() {
		os.Args = oldArgs

		osArgsMu.Unlock()
	}()

	runner()
}

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
	withArgs(t, []string{"ksail"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() should not panic")
	})
}

func TestRunWithHelp(t *testing.T) {
	t.Parallel()

	// Test that run function handles help flag without panicking
	withArgs(t, []string{"ksail", "--help"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with help should not panic")
	})
}

func TestRunWithInvalidCommand(t *testing.T) {
	t.Parallel()

	// Test that run function handles invalid commands without panicking
	withArgs(t, []string{"ksail", "invalid-command"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with invalid command should not panic")
	})
}

func TestRunWithVersionFlag(t *testing.T) {
	t.Parallel()

	// Test that run function handles version flag without panicking
	withArgs(t, []string{"ksail", "--version"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with version should not panic")
	})
}

func TestRunWithSubcommandHelp(t *testing.T) {
	t.Parallel()

	// Test that run function handles init help without panicking
	withArgs(t, []string{"ksail", "init", "--help"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with init help should not panic")
	})
}

func TestMainFunction(t *testing.T) {
	t.Parallel()

	// Test that main() doesn't panic when called
	// Note: We can't easily test main() directly as it may call os.Exit()
	// but we can verify it doesn't panic when run is called properly
	assert.NotPanics(t, func() {
		withArgs(t, []string{"ksail", "--help"}, func() {
			run()
		})
	}, "main() simulation should not panic")
}
