package errorhandler

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
)

// Normalizer transforms raw stderr output captured from Cobra into a final error message.
type Normalizer interface {
	Normalize(raw string) string
}

// Option configures an Executor.
type Option func(*Executor)

// WithNormalizer overrides the default message normalizer used by the executor.
func WithNormalizer(normalizer Normalizer) Option {
	return func(e *Executor) {
		if normalizer != nil {
			e.normalizer = normalizer
		}
	}
}

// Executor coordinates Cobra execution, capturing stderr output and surfacing aggregated errors.
type Executor struct {
	normalizer Normalizer
}

// NewExecutor constructs an Executor with optional functional options.
func NewExecutor(opts ...Option) *Executor {
	executor := &Executor{normalizer: DefaultNormalizer{}}

	for _, opt := range opts {
		opt(executor)
	}

	return executor
}

// Execute runs the provided command while intercepting Cobra's error stream.
// It returns nil on success, or a *CommandError containing both the normalized message
// and the original error to preserve error-chain semantics.
func (e *Executor) Execute(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}

	var errBuf bytes.Buffer

	originalErrWriter := cmd.ErrOrStderr()

	cmd.SetErr(&errBuf)
	defer cmd.SetErr(originalErrWriter)

	err := cmd.Execute()
	if err == nil {
		return nil
	}

	message := ""
	if e.normalizer != nil {
		message = e.normalizer.Normalize(errBuf.String())
	}

	return &CommandError{
		message: message,
		cause:   err,
	}
}

// CommandError represents a Cobra execution failure augmented with normalized stderr output.
type CommandError struct {
	message string
	cause   error
}

// NewCommandError constructs a CommandError with the provided message and cause.
func NewCommandError(message string, cause error) *CommandError {
	return &CommandError{
		message: message,
		cause:   cause,
	}
}

// Error implements the error interface.
func (e *CommandError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.message != "" && e.cause != nil:
		if strings.Contains(e.message, e.cause.Error()) {
			return e.message
		}

		return e.message + ": " + e.cause.Error()
	case e.message != "":
		return e.message
	case e.cause != nil:
		return e.cause.Error()
	default:
		return ""
	}
}

// Unwrap exposes the underlying cause for errors.Is/errors.As consumers.
func (e *CommandError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

// DefaultNormalizer implements Normalizer with the same semantics previously embedded in root.go.
type DefaultNormalizer struct{}

// Normalize trims whitespace, removes redundant "Error:" prefixes, and preserves multi-line usage hints.
func (DefaultNormalizer) Normalize(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return ""
	}

	first := strings.TrimSpace(lines[0])
	first = strings.TrimPrefix(first, "Error: ")
	lines[0] = first

	return strings.Join(lines, "\n")
}
