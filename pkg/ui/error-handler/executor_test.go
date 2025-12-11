package errorhandler_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	errorhandler "github.com/devantler-tech/ksail-go/pkg/ui/error-handler"
)

var (
	errTestBoom        = errors.New("boom")
	errOriginalFailure = errors.New("original failure")
	errBoomOriginal    = errors.New("boom: original failure")
	errJustCause       = errors.New("just cause")
	errWrapped         = errors.New("wrapped")
)

type capturingNormalizer struct {
	message string
	called  bool
}

func (n *capturingNormalizer) Normalize(_ string) string {
	n.called = true

	return n.message
}

func TestExecutorExecuteSuccess(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	executor := errorhandler.NewExecutor()

	err := executor.Execute(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecutorExecuteNilCommand(t *testing.T) {
	t.Parallel()

	executor := errorhandler.NewExecutor()

	err := executor.Execute(nil)
	if err != nil {
		t.Fatalf("expected nil command to succeed, got %v", err)
	}
}

func TestExecutorExecuteInvalidSubcommand(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "test"}
	root.AddCommand(&cobra.Command{Use: "valid"})
	root.SetArgs([]string{"invalid"})

	executor := errorhandler.NewExecutor()

	err := executor.Execute(root)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "unknown command \"invalid\" for \"test\"") {
		t.Fatalf("expected error message to contain unknown command text, got %q", message)
	}

	if strings.Contains(message, "Error: ") {
		t.Fatalf("expected message to strip 'Error:' prefix, got %q", message)
	}

	if !strings.Contains(message, "Run 'test --help' for usage.") {
		t.Fatalf("expected usage hint to be preserved, got %q", message)
	}
}

func TestExecutorWithCustomNormalizer(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errTestBoom
		},
	}

	normalizer := &capturingNormalizer{message: "custom boom"}

	executor := errorhandler.NewExecutor(errorhandler.WithNormalizer(normalizer))

	err := executor.Execute(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !normalizer.called {
		t.Fatal("expected normalizer to be invoked")
	}

	if !strings.HasPrefix(err.Error(), "custom boom") {
		t.Fatalf("expected custom normalized prefix, got %q", err.Error())
	}
}

func TestCommandErrorError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      *errorhandler.CommandError
		expected string
	}{
		{
			name:     "nil receiver returns empty string",
			err:      nil,
			expected: "",
		},
		{
			name:     "message and cause concatenated when distinct",
			err:      errorhandler.NewCommandError("normalized", errOriginalFailure),
			expected: "normalized: original failure",
		},
		{
			name:     "message retained when already includes cause",
			err:      errorhandler.NewCommandError("boom: original failure", errBoomOriginal),
			expected: "boom: original failure",
		},
		{
			name:     "message only when cause missing",
			err:      errorhandler.NewCommandError("only message", nil),
			expected: "only message",
		},
		{
			name:     "cause only when message empty",
			err:      errorhandler.NewCommandError("", errJustCause),
			expected: "just cause",
		},
		{
			name:     "empty struct returns empty string",
			err:      &errorhandler.CommandError{},
			expected: "",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := commandErrorString(testCase.err)

			if actual != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}

func TestCommandErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errWrapped
	err := errorhandler.NewCommandError("", cause)

	if !errors.Is(err.Unwrap(), cause) {
		t.Fatalf("expected unwrap to return original cause")
	}

	if (*errorhandler.CommandError)(nil).Unwrap() != nil {
		t.Fatalf("expected nil receiver unwrap to return nil")
	}
}

func TestDefaultNormalizerNormalize(t *testing.T) {
	t.Parallel()

	normalizer := errorhandler.DefaultNormalizer{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input returns empty string",
			input:    "   \n\t  ",
			expected: "",
		},
		{
			name:     "strips error prefix and trims",
			input:    "  Error: something bad \nRun help\n",
			expected: "something bad\nRun help",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := normalizer.Normalize(testCase.input)
			if actual != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}

func commandErrorString(err *errorhandler.CommandError) string {
	if err == nil {
		var cmdErr *errorhandler.CommandError

		return cmdErr.Error()
	}

	return err.Error()
}
