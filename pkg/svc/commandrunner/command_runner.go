// Package commandrunner provides helpers for executing Cobra commands while
// capturing their output and displaying it to the console.
package commandrunner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CommandResult captures the stdout and stderr collected during a Cobra command
// execution.
type CommandResult struct {
	Stdout string
	Stderr string
}

// CommandRunner executes Cobra commands while capturing their output.
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
// It displays output to the console in real-time (like running the binary directly)
// while also capturing output for programmatic use.
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

// Run executes a Cobra command and displays output in real-time to console.
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
