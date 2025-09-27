package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // Shared across tests to synchronize os.Args mutations.
var (
	osArgsMu    sync.Mutex
	osArgsStack [][]string
)

func withArgs(t *testing.T, args []string, runner func()) {
	t.Helper()

	nextArgs := append([]string(nil), args...)

	osArgsMu.Lock()
	snapshot := append([]string(nil), os.Args...)
	osArgsStack = append(osArgsStack, snapshot)
	os.Args = nextArgs
	osArgsMu.Unlock()

	defer func() {
		osArgsMu.Lock()

		stackLen := len(osArgsStack)
		if stackLen == 0 {
			os.Args = nil

			osArgsMu.Unlock()

			t.Fatalf("withArgs stack underflow; this indicates an imbalance in helper usage")

			return
		}

		previous := osArgsStack[stackLen-1]
		osArgsStack = osArgsStack[:stackLen-1]
		os.Args = previous

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

func TestWithArgsNested(t *testing.T) {
	t.Parallel()

	withArgs(t, []string{"ksail", "outer"}, func() {
		assert.Equal(t, []string{"ksail", "outer"}, os.Args)

		withArgs(t, []string{"ksail", "inner"}, func() {
			assert.Equal(t, []string{"ksail", "inner"}, os.Args)
		})

		assert.Equal(t, []string{"ksail", "outer"}, os.Args)
	})
}

func TestWithArgsTableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "root", args: []string{"ksail"}},
		{name: "help", args: []string{"ksail", "--help"}},
		{name: "version", args: []string{"ksail", "--version"}},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			withArgs(t, tt.args, func() {
				assert.Equal(t, tt.args, os.Args)
			})
		})
	}
}

func TestWithArgsPreservesPanicsAndStackTrace(t *testing.T) {
	t.Parallel()

	var (
		recovered any
		stack     []byte
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = r
				stack = debug.Stack()
			}
		}()

		withArgs(t, []string{"ksail", "panic"}, func() {
			panic("test panic")
		})
	}()

	require.NotNil(t, recovered)
	require.Contains(t, fmt.Sprint(recovered), "test panic")
	require.NotEmpty(t, stack)
	assert.Contains(t, string(stack), "TestWithArgsPreservesPanicsAndStackTrace")
}
