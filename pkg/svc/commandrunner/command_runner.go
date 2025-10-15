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

	originalStdout := os.Stdout
	originalStderr := os.Stderr
	originalExit := k3dlog.Log().ExitFunc

	stdoutDest := io.Writer(&stdoutBuf)
	if originalStdout != nil {
		stdoutDest = io.MultiWriter(&stdoutBuf, originalStdout)
	}

	stderrDest := io.Writer(&stderrBuf)
	if originalStderr != nil {
		stderrDest = io.MultiWriter(&stderrBuf, originalStderr)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(stdoutDest, stdoutReader)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(stderrDest, stderrReader)
	}()

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
			err = MergeCommandError(fatalErr, res)
		case err != nil:
			err = MergeCommandError(err, res)
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

	logger := k3dlog.Log()
	savedHooks := cloneHooks(logger.Hooks)
	workingHooks := stripStdoutInfoHooks(cloneHooks(logger.Hooks), originalStdout)
	originalLoggerOut := logger.Out
	logger.ReplaceHooks(workingHooks)
	logger.SetOutput(io.Discard)
	formatter := cloneFormatter(logger.Formatter, originalStdout)
	logger.AddHook(&pipeForwardHook{writer: stdoutWriter, formatter: formatter})

	defer func() {
		logger.SetOutput(originalLoggerOut)
		logger.ReplaceHooks(savedHooks)
		logger.ExitFunc = originalExit
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	cmd.SetOut(stdoutWriter)
	cmd.SetErr(stderrWriter)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	logger.ExitFunc = func(code int) {
		fatalErr = fmt.Errorf("k3d command exited with status %d", code)
		panic(exitSentinel{})
	}

	err = cmd.ExecuteContext(ctx)
	return res, err
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
		return err
	}
	if _, err = h.writer.Write(formatted); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			return nil
		}
		return err
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
		}
	}

	return false
}

func cloneFormatter(base logrus.Formatter, stdout *os.File) logrus.Formatter {
	if base == nil {
		return &logrus.TextFormatter{ForceColors: stdout != nil}
	}

	if tf, ok := base.(*logrus.TextFormatter); ok {
		formatter := &logrus.TextFormatter{
			ForceColors:               tf.ForceColors || stdout != nil,
			DisableColors:             tf.DisableColors,
			ForceQuote:                tf.ForceQuote,
			DisableQuote:              tf.DisableQuote,
			EnvironmentOverrideColors: tf.EnvironmentOverrideColors,
			DisableTimestamp:          tf.DisableTimestamp,
			FullTimestamp:             tf.FullTimestamp,
			TimestampFormat:           tf.TimestampFormat,
			DisableSorting:            tf.DisableSorting,
			SortingFunc:               tf.SortingFunc,
			DisableLevelTruncation:    tf.DisableLevelTruncation,
			PadLevelText:              tf.PadLevelText,
			QuoteEmptyFields:          tf.QuoteEmptyFields,
			FieldMap:                  tf.FieldMap,
			CallerPrettyfier:          tf.CallerPrettyfier,
		}
		return formatter
	}

	return base
}
