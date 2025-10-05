package errorhandler

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestExecutorExecuteSuccess(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	executor := NewExecutor()

	err := executor.Execute(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecutorExecuteNilCommand(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()

	if err := executor.Execute(nil); err != nil {
		t.Fatalf("expected nil command to succeed, got %v", err)
	}
}

func TestExecutorExecuteInvalidSubcommand(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "test"}
	root.AddCommand(&cobra.Command{Use: "valid"})
	root.SetArgs([]string{"invalid"})

	executor := NewExecutor()

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

var errTestBoom = errors.New("boom")

func TestExecutorWithCustomNormalizer(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errTestBoom
		},
	}

	normalizer := NewMockNormalizer(t)
	normalizer.EXPECT().Normalize(mock.Anything).Return("custom boom")

	executor := NewExecutor(WithNormalizer(normalizer))

	err := executor.Execute(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.HasPrefix(err.Error(), "custom boom") {
		t.Fatalf("expected custom normalized prefix, got %q", err.Error())
	}
}

func TestCommandErrorError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      *CommandError
		expected string
	}{
		{
			name:     "nil receiver returns empty string",
			err:      nil,
			expected: "",
		},
		{
			name: "message and cause concatenated when distinct",
			err: &CommandError{
				message: "normalized",
				cause:   errors.New("original failure"),
			},
			expected: "normalized: original failure",
		},
		{
			name: "message retained when already includes cause",
			err: &CommandError{
				message: "boom: original failure",
				cause:   errors.New("boom: original failure"),
			},
			expected: "boom: original failure",
		},
		{
			name: "message only when cause missing",
			err: &CommandError{
				message: "only message",
			},
			expected: "only message",
		},
		{
			name: "cause only when message empty",
			err: &CommandError{
				cause: errors.New("just cause"),
			},
			expected: "just cause",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.err.Error()
			if actual != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func TestCommandErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("wrapped")
	err := &CommandError{cause: cause}

	if err.Unwrap() != cause {
		t.Fatalf("expected unwrap to return original cause")
	}

	if (*CommandError)(nil).Unwrap() != nil {
		t.Fatalf("expected nil receiver unwrap to return nil")
	}
}

func TestDefaultNormalizerNormalize(t *testing.T) {
	t.Parallel()

	normalizer := DefaultNormalizer{}

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

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := normalizer.Normalize(tc.input)
			if actual != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
