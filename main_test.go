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

// osArgsMu guards temporary mutations to os.Args during tests.
//
//nolint:gochecknoglobals // Single shared mutex required to serialize os.Args usage.
var osArgsMu sync.Mutex

// withOSArgs temporarily replaces os.Args for the duration of runner.
// This helper is test-only and intentionally non-reentrant; invoking tests
// must avoid nested usage to keep the implementation simple and portable.
func withOSArgs(t *testing.T, args []string, runner func()) {
	t.Helper()

	osArgsMu.Lock()
	defer osArgsMu.Unlock()

	previousArgs := append([]string(nil), os.Args...)
	os.Args = append([]string(nil), args...)

	defer func() {
		os.Args = previousArgs
	}()

	runner()
}

func TestVersionVariables(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "dev", version)
	assert.Equal(t, "none", commit)
	assert.Equal(t, "unknown", date)
}

func TestRunBasic(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		assert.Equal(t, 0, runWithArgs(nil))
	})
}

func TestRunWithHelp(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		assert.Equal(t, 0, runWithArgs([]string{"--help"}))
	}, "run() with help should not panic")
}

func TestRunWithInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		assert.Equal(t, 1, runWithArgs([]string{"invalid-command"}))
	}, "run() with invalid command should not panic")
}

func TestRunWithVersionFlag(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		assert.Equal(t, 0, runWithArgs([]string{"--version"}))
	}, "run() with version should not panic")
}

func TestRunWithSubcommandHelp(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		assert.Equal(t, 0, runWithArgs([]string{"init", "--help"}))
	}, "run() with init help should not panic")
}

func TestMainFunction(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		withOSArgs(t, []string{"ksail", "--help"}, func() {
			run()
		})
	}, "main() simulation should not panic")
}

func TestMainExecutesWithoutExit(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		withOSArgs(t, []string{"ksail"}, func() {
			main()
		})
	})
}

func TestRunSafelyReturnsZeroOnSuccess(t *testing.T) {
	t.Parallel()

	exitCode := runSafelyWithArgs(nil)

	assert.Equal(t, 0, exitCode)
}

func TestRunSafelyReturnsErrorCode(t *testing.T) {
	t.Parallel()

	exitCode := runSafelyWithArgs([]string{"invalid-command"})

	assert.Equal(t, 1, exitCode)
}

func TestWithOSArgsTableDriven(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			withOSArgs(t, tt.args, func() {
				assert.Equal(t, tt.args, os.Args)
			})
		})
	}
}

func TestWithOSArgsPreservesPanicsAndStackTrace(t *testing.T) {
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

		withOSArgs(t, []string{"ksail", "panic"}, func() {
			panic("test panic")
		})
	}()

	require.NotNil(t, recovered)
	require.Contains(t, fmt.Sprint(recovered), "test panic")
	require.NotEmpty(t, stack)
	assert.Contains(t, string(stack), "TestWithOSArgsPreservesPanicsAndStackTrace")
}
