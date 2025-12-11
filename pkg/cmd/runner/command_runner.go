package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CommandResult captures the stdout and stderr collected during a Cobra command execution.
// Both fields contain the complete output from the command, including any output produced
// before an error occurred.
type CommandResult struct {
	Stdout string
	Stderr string
}

// CommandRunner executes Cobra commands while capturing their output.
// Implementations should display output to stdout/stderr in real-time while also
// capturing it for programmatic access via CommandResult.
type CommandRunner interface {
	Run(ctx context.Context, cmd *cobra.Command, args []string) (CommandResult, error)
}

// CobraCommandRunner executes any Cobra command with console output.
// This runner displays command output to stdout/stderr in real-time while
// also capturing it for the result.
type CobraCommandRunner struct {
	stdout io.Writer
	stderr io.Writer
}

// NewCobraCommandRunner creates a command runner that works with any Cobra command.
// It displays output to stdout/stderr in real-time (like running the binary directly)
// while also capturing output for programmatic use in the CommandResult.
//
// If stdout or stderr are nil, they default to os.Stdout and os.Stderr respectively.
func NewCobraCommandRunner(stdout, stderr io.Writer) *CobraCommandRunner {
	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	return &CobraCommandRunner{
		stdout: stdout,
		stderr: stderr,
	}
}

// Run executes a Cobra command and displays output in real-time to the console.
// The command's output streams are configured to write to both capture buffers and
// the configured stdout/stderr writers, providing the same behavior as running the
// binary directly while also making the output available programmatically.
//
// The command is executed with the provided context and arguments. Usage and error
// messages are silenced since this runner handles error reporting.
//
// Returns the captured output and any error from command execution.
func (r *CobraCommandRunner) Run(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
) (CommandResult, error) {
	var outBuf, errBuf bytes.Buffer

	// Use io.MultiWriter to display AND capture output
	// This provides the same behavior as running the binary directly
	cmd.SetOut(io.MultiWriter(&outBuf, r.stdout))
	cmd.SetErr(io.MultiWriter(&errBuf, r.stderr))

	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	execErr := cmd.ExecuteContext(ctx)
	if execErr != nil {
		return CommandResult{
			Stdout: outBuf.String(),
			Stderr: errBuf.String(),
		}, fmt.Errorf("command execution failed: %w", execErr)
	}

	return CommandResult{
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
	}, nil
}

// MergeCommandError enriches a base error with captured stdout/stderr when available.
// This is useful for providing detailed error messages that include command output.
//
// If the base error is nil, returns nil. If there is no captured output (both stdout
// and stderr are empty after trimming whitespace), returns the base error unchanged.
// Otherwise, returns a new error that wraps the base error and appends the captured
// output, with stderr appearing before stdout.
func MergeCommandError(base error, res CommandResult) error {
	if base == nil {
		return nil
	}

	var details []string
	if trimmed := strings.TrimSpace(res.Stderr); trimmed != "" {
		details = append(details, trimmed)
	}

	if trimmed := strings.TrimSpace(res.Stdout); trimmed != "" {
		details = append(details, trimmed)
	}

	if len(details) == 0 {
		return base
	}

	return fmt.Errorf("%w: %s", base, strings.Join(details, " | "))
}
