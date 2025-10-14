package commandrunner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	k3dlog "github.com/k3d-io/k3d/v5/pkg/logger"
	"github.com/spf13/cobra"
)

type CommandResult struct {
	Stdout string
	Stderr string
}

type CommandRunner interface {
	Run(ctx context.Context, cmd *cobra.Command, args []string) (CommandResult, error)
}

type cobraCommandRunner struct{}

func NewCobraCommandRunner() CommandRunner {
	return &cobraCommandRunner{}
}

func (r *cobraCommandRunner) Run(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
) (res CommandResult, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	stdoutReader, stdoutWriter, pipeErr := os.Pipe()
	if pipeErr != nil {
		return res, fmt.Errorf("capture stdout: %w", pipeErr)
	}

	stderrReader, stderrWriter, pipeErr := os.Pipe()
	if pipeErr != nil {
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()
		return res, fmt.Errorf("capture stderr: %w", pipeErr)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdoutReader)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stderrBuf, stderrReader)
	}()

	originalStdout := os.Stdout
	originalStderr := os.Stderr
	originalLoggerOut := k3dlog.Log().Out
	originalExit := k3dlog.Log().ExitFunc

	type exitSentinel struct{}
	var fatalErr error

	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(exitSentinel); !ok {
				panic(recovered)
			}
		}

		switch {
		case fatalErr != nil:
			err = mergeCommandError(fatalErr, res)
		case err != nil:
			err = mergeCommandError(err, res)
		}
	}()

	defer func() {
		_ = stdoutWriter.Close()
		_ = stderrWriter.Close()
		wg.Wait()
		_ = stdoutReader.Close()
		_ = stderrReader.Close()

		res.Stdout = stdoutBuf.String()
		res.Stderr = stderrBuf.String()
	}()

	defer func() {
		k3dlog.Log().SetOutput(originalLoggerOut)
		k3dlog.Log().ExitFunc = originalExit
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	k3dlog.Log().SetOutput(stdoutWriter)

	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	cmd.SetOut(stdoutWriter)
	cmd.SetErr(stderrWriter)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	k3dlog.Log().ExitFunc = func(code int) {
		fatalErr = fmt.Errorf("k3d command exited with status %d", code)
		panic(exitSentinel{})
	}

	err = cmd.ExecuteContext(ctx)
	return res, err
}

func mergeCommandError(base error, res CommandResult) error {
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
