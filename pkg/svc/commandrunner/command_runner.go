// Package commandrunner provides helpers for executing Cobra commands while
// capturing their output and translating k3d logging semantics.
package commandrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	k3dlog "github.com/k3d-io/k3d/v5/pkg/logger"
	"github.com/sirupsen/logrus"
	logwriter "github.com/sirupsen/logrus/hooks/writer"
	"github.com/spf13/cobra"
)

const streamCopyWorkers = 2

var (
	errK3dCommandExit    = errors.New("commandrunner: k3d command exited")
	errLoggerUnavailable = errors.New("commandrunner: logger not available")
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

// CobraCommandRunner executes Cobra commands while mirroring k3d logging semantics.
type CobraCommandRunner struct{}

// NewCobraCommandRunner creates a command runner that wraps Cobra execution
// with stdout/stderr capture compatible with k3d's logging behavior.
func NewCobraCommandRunner() *CobraCommandRunner {
	return &CobraCommandRunner{}
}

// Run executes the provided Cobra command while capturing stdout/stderr.
func (r *CobraCommandRunner) Run(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
) (CommandResult, error) {
	execCtx, err := newCommandExecution()
	if err != nil {
		return CommandResult{}, fmt.Errorf("prepare command execution: %w", err)
	}

	if err = execCtx.configureLogger(); err != nil {
		execCtx.restore()

		return CommandResult{}, fmt.Errorf("configure logger: %w", err)
	}
	defer execCtx.resetLogger()

	execCtx.prepareCommand(ctx, cmd, args)

	if err = execCtx.execute(ctx, cmd); err != nil {
		execCtx.restore()
		result := execCtx.result()

		return result, MergeCommandError(err, result)
	}

	if execCtx.fatalErr != nil {
		execCtx.restore()
		result := execCtx.result()

		return result, MergeCommandError(execCtx.fatalErr, result)
	}

	execCtx.restore()

	return execCtx.result(), nil
}

type commandExecution struct {
	stdoutBuffer bytes.Buffer
	stderrBuffer bytes.Buffer

	stdoutReader *os.File
	stdoutWriter *os.File
	stderrReader *os.File
	stderrWriter *os.File

	originalStdout *os.File
	originalStderr *os.File

	stdoutDest io.Writer
	stderrDest io.Writer

	waitGroup sync.WaitGroup

	logger            *logrus.Logger
	savedHooks        logrus.LevelHooks
	originalLoggerOut io.Writer
	originalExit      func(int)
	formatter         logrus.Formatter

	fatalErr error
}

func newCommandExecution() (*commandExecution, error) {
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("capture stdout: %w", err)
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()

		return nil, fmt.Errorf("capture stderr: %w", err)
	}

	ctx := &commandExecution{
		stdoutReader:      stdoutReader,
		stdoutWriter:      stdoutWriter,
		stderrReader:      stderrReader,
		stderrWriter:      stderrWriter,
		originalStdout:    os.Stdout,
		originalStderr:    os.Stderr,
		logger:            k3dlog.Log(),
		originalLoggerOut: io.Discard, // placeholder, overwritten in configureLogger.
	}

	ctx.stdoutDest = ctx.buildDestWriter(ctx.originalStdout, &ctx.stdoutBuffer)
	ctx.stderrDest = ctx.buildDestWriter(ctx.originalStderr, &ctx.stderrBuffer)

	ctx.waitGroup.Add(streamCopyWorkers)

	go ctx.copyStream(ctx.stdoutReader, ctx.stdoutDest)
	go ctx.copyStream(ctx.stderrReader, ctx.stderrDest)

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	return ctx, nil
}

func (c *commandExecution) buildDestWriter(original *os.File, buffer *bytes.Buffer) io.Writer {
	if original == nil {
		return buffer
	}

	return io.MultiWriter(buffer, original)
}

func (c *commandExecution) copyStream(reader *os.File, writer io.Writer) {
	defer c.waitGroup.Done()

	_, _ = io.Copy(writer, reader)
}

func (c *commandExecution) configureLogger() error {
	if c.logger == nil {
		return errLoggerUnavailable
	}

	c.savedHooks = cloneHooks(c.logger.Hooks)
	workingHooks := stripStdoutInfoHooks(cloneHooks(c.logger.Hooks), c.originalStdout)
	c.originalLoggerOut = c.logger.Out
	c.logger.ReplaceHooks(workingHooks)
	c.logger.SetOutput(io.Discard)

	switch typedFormatter := c.logger.Formatter.(type) {
	case *logrus.TextFormatter:
		c.formatter = &logrus.TextFormatter{
			ForceColors:               typedFormatter.ForceColors || c.originalStdout != nil,
			DisableColors:             typedFormatter.DisableColors,
			ForceQuote:                typedFormatter.ForceQuote,
			DisableQuote:              typedFormatter.DisableQuote,
			EnvironmentOverrideColors: typedFormatter.EnvironmentOverrideColors,
			DisableTimestamp:          typedFormatter.DisableTimestamp,
			FullTimestamp:             typedFormatter.FullTimestamp,
			TimestampFormat:           typedFormatter.TimestampFormat,
			DisableSorting:            typedFormatter.DisableSorting,
			SortingFunc:               typedFormatter.SortingFunc,
			DisableLevelTruncation:    typedFormatter.DisableLevelTruncation,
			PadLevelText:              typedFormatter.PadLevelText,
			QuoteEmptyFields:          typedFormatter.QuoteEmptyFields,
			FieldMap:                  typedFormatter.FieldMap,
			CallerPrettyfier:          typedFormatter.CallerPrettyfier,
		}
	case nil:
		c.formatter = &logrus.TextFormatter{ForceColors: c.originalStdout != nil}
	default:
		c.formatter = typedFormatter
	}

	c.logger.AddHook(&pipeForwardHook{writer: c.stdoutWriter, formatter: c.formatter})

	c.originalExit = c.logger.ExitFunc

	type exitSentinel struct{}

	c.logger.ExitFunc = func(code int) {
		c.fatalErr = fmt.Errorf("%w: status %d", errK3dCommandExit, code)

		panic(exitSentinel{})
	}

	return nil
}

func (c *commandExecution) prepareCommand(ctx context.Context, cmd *cobra.Command, args []string) {
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	cmd.SetOut(c.stdoutWriter)
	cmd.SetErr(c.stderrWriter)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
}

func (c *commandExecution) execute(ctx context.Context, cmd *cobra.Command) error {
	type exitSentinel struct{}

	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(exitSentinel); !ok {
				panic(recovered)
			}
		}
	}()

	execErr := cmd.ExecuteContext(ctx)
	if execErr != nil {
		return fmt.Errorf("execute command: %w", execErr)
	}

	return nil
}

func (c *commandExecution) resetLogger() {
	if c.logger == nil {
		return
	}

	c.logger.SetOutput(c.originalLoggerOut)
	c.logger.ReplaceHooks(c.savedHooks)
	c.logger.ExitFunc = c.originalExit
}

func (c *commandExecution) restore() {
	_ = c.stdoutWriter.Close()
	_ = c.stderrWriter.Close()

	c.waitGroup.Wait()

	_ = c.stdoutReader.Close()
	_ = c.stderrReader.Close()

	os.Stdout = c.originalStdout
	os.Stderr = c.originalStderr
}

func (c *commandExecution) result() CommandResult {
	return CommandResult{
		Stdout: c.stdoutBuffer.String(),
		Stderr: c.stderrBuffer.String(),
	}
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

type pipeForwardHook struct {
	writer    io.Writer
	formatter logrus.Formatter
}

func (h *pipeForwardHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *pipeForwardHook) Fire(entry *logrus.Entry) error {
	if h.writer == nil || h.formatter == nil {
		return nil
	}

	dup := entry.Dup()
	dup.Level = entry.Level
	dup.Message = entry.Message

	formatted, err := h.formatter.Format(dup)
	if err != nil {
		return fmt.Errorf("format log entry: %w", err)
	}

	if _, err = h.writer.Write(formatted); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			return nil
		}

		return fmt.Errorf("write log entry: %w", err)
	}

	return nil
}

func cloneHooks(hooks logrus.LevelHooks) logrus.LevelHooks {
	if hooks == nil {
		return nil
	}

	cloned := make(logrus.LevelHooks, len(hooks))
	for level, levelHooks := range hooks {
		if len(levelHooks) == 0 {
			continue
		}

		copies := make([]logrus.Hook, len(levelHooks))
		copy(copies, levelHooks)
		cloned[level] = copies
	}

	return cloned
}

func stripStdoutInfoHooks(hooks logrus.LevelHooks, stdout *os.File) logrus.LevelHooks {
	if hooks == nil {
		return nil
	}

	filtered := make(logrus.LevelHooks, len(hooks))
	for level, levelHooks := range hooks {
		var kept []logrus.Hook

		for _, hook := range levelHooks {
			if isStdoutInfoWriterHook(hook, stdout) {
				continue
			}

			kept = append(kept, hook)
		}

		if len(kept) > 0 {
			filtered[level] = kept
		}
	}

	return filtered
}

func isStdoutInfoWriterHook(hook logrus.Hook, stdout *os.File) bool {
	if stdout == nil {
		return false
	}

	writerHook, ok := hook.(*logwriter.Hook)
	if !ok {
		return false
	}

	if writerHook.Writer != stdout {
		return false
	}

	for _, level := range writerHook.LogLevels {
		switch level {
		case logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel:
			return true
		case logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			continue
		default:
			continue
		}
	}

	return false
}
