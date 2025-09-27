package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionVariables(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "dev", version)
	assert.Equal(t, "none", commit)
	assert.Equal(t, "unknown", date)
}

func TestRunWithArgsScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "root", want: 0},
		{name: "help", args: []string{"--help"}, want: 0},
		{name: "version", args: []string{"--version"}, want: 0},
		{name: "invalid", args: []string{"invalid-command"}, want: 1},
		{name: "init-help", args: []string{"init", "--help"}, want: 0},
	}

	for i := range tests {
		testCase := tests[i]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.want, runWithArgs(testCase.args))
		})
	}
}

func TestRunSafelyRecoversFromPanic(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	exitCode := runSafely(nil, func([]string) int {
		panic("test panic")
	}, &output)

	assert.Equal(t, 1, exitCode)
	require.Contains(t, output.String(), "test panic")
	require.Contains(t, output.String(), "TestRunSafelyRecoversFromPanic")
}

func TestRunSafelyExecutesRunWithArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "success", want: 0},
		{name: "error", args: []string{"invalid-command"}, want: 1},
	}

	for i := range tests {
		testCase := tests[i]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var errOutput bytes.Buffer

			assert.Equal(t, testCase.want, runSafely(testCase.args, runWithArgs, &errOutput))
		})
	}
}

func TestRunSafelyPropagatesRunnerExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		runner func([]string) int
		want   int
	}{
		{
			name:   "success",
			runner: func([]string) int { return 0 },
			want:   0,
		},
		{
			name:   "failure",
			runner: func([]string) int { return 2 },
			want:   2,
		},
	}

	for i := range tests {
		testCase := tests[i]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			assert.Equal(t, testCase.want, runSafely(nil, testCase.runner, &output))
		})
	}
}
