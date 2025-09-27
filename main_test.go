package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // Shared across tests to synchronize os.Args mutations.
var (
	osArgsLock  reentrantMutex
	osArgsStack [][]string
	commandMu   sync.Mutex
)

type reentrantMutex struct {
	mu    sync.Mutex
	owner uint64
	depth int
}

func (m *reentrantMutex) Lock() {
	gid := goroutineID()

	if atomic.LoadUint64(&m.owner) == gid {
		m.depth++

		return
	}

	m.mu.Lock()
	atomic.StoreUint64(&m.owner, gid)
	m.depth = 1
}

func (m *reentrantMutex) Unlock() {
	gid := goroutineID()

	if atomic.LoadUint64(&m.owner) != gid {
		panic("reentrantMutex: unlock attempted by non-owner goroutine")
	}

	m.depth--
	if m.depth == 0 {
		atomic.StoreUint64(&m.owner, 0)
		m.mu.Unlock()
	}
}

func goroutineID() uint64 {
	var buf [64]byte

	n := runtime.Stack(buf[:], false)

	fields := bytes.Fields(buf[:n])
	if len(fields) < 2 {
		return 0
	}

	id, err := strconv.ParseUint(string(fields[1]), 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func withArgs(t *testing.T, args []string, runner func()) {
	t.Helper()

	clonedArgs := append([]string(nil), args...)

	osArgsLock.Lock()

	previousArgs := append([]string(nil), os.Args...)
	osArgsStack = append(osArgsStack, previousArgs)
	os.Args = clonedArgs

	defer func() {
		defer osArgsLock.Unlock()

		stackLen := len(osArgsStack)
		if stackLen == 0 {
			os.Args = nil

			t.Fatalf("withArgs stack underflow; this indicates an imbalance in helper usage")

			return
		}

		previous := osArgsStack[stackLen-1]
		osArgsStack = osArgsStack[:stackLen-1]
		os.Args = previous
	}()

	runner()
}

func withCommand(t *testing.T, args []string, runner func()) {
	t.Helper()

	withArgs(t, args, func() {
		commandMu.Lock()
		defer commandMu.Unlock()

		runner()
	})
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
	withCommand(t, []string{"ksail"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() should not panic")
	})
}

func TestRunWithHelp(t *testing.T) {
	t.Parallel()

	// Test that run function handles help flag without panicking
	withCommand(t, []string{"ksail", "--help"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with help should not panic")
	})
}

func TestRunWithInvalidCommand(t *testing.T) {
	t.Parallel()

	// Test that run function handles invalid commands without panicking
	withCommand(t, []string{"ksail", "invalid-command"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with invalid command should not panic")
	})
}

func TestRunWithVersionFlag(t *testing.T) {
	t.Parallel()

	// Test that run function handles version flag without panicking
	withCommand(t, []string{"ksail", "--version"}, func() {
		assert.NotPanics(t, func() {
			run()
		}, "run() with version should not panic")
	})
}

func TestRunWithSubcommandHelp(t *testing.T) {
	t.Parallel()

	// Test that run function handles init help without panicking
	withCommand(t, []string{"ksail", "init", "--help"}, func() {
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
		withCommand(t, []string{"ksail", "--help"}, func() {
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
