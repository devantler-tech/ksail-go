package commandrunner //nolint:testpackage // Access internal helpers to increase coverage.

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logwriter "github.com/sirupsen/logrus/hooks/writer"
	"github.com/spf13/cobra"
)

var (
	errBaseFailure   = errors.New("base failure")
	errFormatFailed  = errors.New("format failed")
	errWriteFailed   = errors.New("write failed")
	errCommandFailed = errors.New("boom")
	errBaseOnly      = errors.New("base error only")
)

//nolint:paralleltest // Serializes stdout manipulation to avoid race with global stdio.
func TestCobraCommandRunner_RunPropagatesStdout(t *testing.T) {
	runner := NewCobraCommandRunner()
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("hello world")
		},
	}

	res, err := runner.Run(context.Background(), cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Stdout, "hello world") {
		t.Fatalf("expected stdout to contain greeting, got %q", res.Stdout)
	}
}

//nolint:paralleltest // Serializes stdout manipulation to avoid race with global stdio.
func TestCobraCommandRunner_RunReturnsMergedError(t *testing.T) {
	runner := NewCobraCommandRunner()
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("info output")
			cmd.PrintErrln("stderr detail")

			return errCommandFailed
		},
	}

	res, err := runner.Run(context.Background(), cmd, nil)
	if err == nil {
		t.Fatal("expected error when command fails")
	}

	msg := err.Error()
	if !strings.Contains(msg, "execute command: boom") {
		t.Fatalf("expected base error in message, got %q", msg)
	}

	if !strings.Contains(msg, "stderr detail | info output") {
		t.Fatalf("expected merged output, got %q", msg)
	}

	if !strings.Contains(res.Stdout, "info output") {
		t.Fatalf("expected stdout capture, got %q", res.Stdout)
	}

	if !strings.Contains(res.Stderr, "stderr detail") {
		t.Fatalf("expected stderr capture, got %q", res.Stderr)
	}
}

func TestMergeCommandError_AppendsStdStreams(t *testing.T) {
	t.Parallel()

	res := CommandResult{Stdout: "info", Stderr: "fail"}

	err := MergeCommandError(errBaseFailure, res)
	if err == nil {
		t.Fatal("expected merged error")
	}

	merged := err.Error()
	if !strings.Contains(merged, errBaseFailure.Error()) {
		t.Fatalf("expected base error in output, got %q", merged)
	}

	if !strings.Contains(merged, "info") || !strings.Contains(merged, "fail") {
		t.Fatalf("expected stdout and stderr in output, got %q", merged)
	}
}

func TestMergeCommandError_NilBaseReturnsNil(t *testing.T) {
	t.Parallel()

	err := MergeCommandError(nil, CommandResult{})
	if err != nil {
		t.Fatalf("expected nil when base error nil, got %v", err)
	}
}

func TestMergeCommandError_NoDetailsReturnsBase(t *testing.T) {
	t.Parallel()

	base := errBaseOnly
	res := CommandResult{Stdout: "\n\t", Stderr: ""}

	merged := MergeCommandError(base, res)
	if !errors.Is(merged, base) {
		t.Fatalf("expected original error when no details, got %v", merged)
	}
}

//nolint:paralleltest // Serializes stdout manipulation to avoid race with global stdio.
func TestCommandExecutionExecuteWrapsError(t *testing.T) {
	execCtx := setupCommandExecution(t)

	err := execCtx.configureLogger()
	if err != nil {
		t.Fatalf("configure logger: %v", err)
	}

	ctx := context.Background()
	cmd := &cobra.Command{
		RunE: func(*cobra.Command, []string) error {
			return errBaseFailure
		},
	}

	execCtx.prepareCommand(ctx, cmd, nil)

	execErr := execCtx.execute(ctx, cmd)
	if execErr == nil {
		t.Fatal("expected execute error")
	}

	if !errors.Is(execErr, errBaseFailure) {
		t.Fatalf("expected wrapped base error, got %v", execErr)
	}

	if !strings.Contains(execErr.Error(), "execute command: base failure") {
		t.Fatalf("expected wrapped message, got %q", execErr.Error())
	}
}

func TestCommandExecutionConfigureLoggerRequiresLogger(t *testing.T) {
	t.Parallel()

	ce := &commandExecution{}

	err := ce.configureLogger()
	if !errors.Is(err, errLoggerUnavailable) {
		t.Fatalf("expected errLoggerUnavailable, got %v", err)
	}
}

//nolint:paralleltest // Mutates shared logger formatter state during configuration.
func TestCommandExecutionConfigureLoggerFormatterSelection(t *testing.T) {
	t.Run("assigns text formatter when logger formatter nil", func(t *testing.T) {
		execCtx := setupCommandExecution(t)

		originalFormatter := execCtx.logger.Formatter
		execCtx.logger.Formatter = nil

		t.Cleanup(func() {
			execCtx.logger.Formatter = originalFormatter
		})

		err := execCtx.configureLogger()
		if err != nil {
			t.Fatalf("configure logger: %v", err)
		}

		formatter, ok := execCtx.formatter.(*logrus.TextFormatter)
		if !ok {
			t.Fatalf("expected text formatter, got %T", execCtx.formatter)
		}

		expectForceColors := execCtx.originalStdout != nil
		if formatter.ForceColors != expectForceColors {
			t.Fatalf("expected ForceColors=%t, got %t", expectForceColors, formatter.ForceColors)
		}
	})

	t.Run("retains custom formatter instance", func(t *testing.T) {
		execCtx := setupCommandExecution(t)

		custom := &logrus.JSONFormatter{}
		previous := execCtx.logger.Formatter
		execCtx.logger.Formatter = custom

		t.Cleanup(func() {
			execCtx.logger.Formatter = previous
		})

		err := execCtx.configureLogger()
		if err != nil {
			t.Fatalf("configure logger: %v", err)
		}

		if execCtx.formatter != custom {
			t.Fatalf("expected custom formatter to be retained, got %T", execCtx.formatter)
		}
	})
}

//nolint:paralleltest // Exercises logger exit function that mutates shared state.
func TestCommandExecutionConfigureLoggerExitFuncCapturesStatus(t *testing.T) {
	execCtx := setupCommandExecution(t)

	err := execCtx.configureLogger()
	if err != nil {
		t.Fatalf("configure logger: %v", err)
	}

	func() {
		defer func() {
			if recovered := recover(); recovered == nil {
				t.Fatal("expected panic from exit func")
			}
		}()

		execCtx.logger.ExitFunc(42)
	}()

	if execCtx.fatalErr == nil {
		t.Fatal("expected fatal error to be recorded")
	}

	if !strings.Contains(execCtx.fatalErr.Error(), "status 42") {
		t.Fatalf("expected status code in fatal error, got %q", execCtx.fatalErr.Error())
	}
}

func TestPipeForwardHookWritesFormattedEntry(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	hook := &pipeForwardHook{
		writer: &buf,
		formatter: stubFormatter{
			output: []byte("formatted\n"),
		},
	}

	entry := newLogEntry(logrus.InfoLevel, "message")

	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != "formatted\n" {
		t.Fatalf("expected formatted output, got %q", buf.String())
	}
}

func TestPipeForwardHookIgnoresClosedPipe(t *testing.T) {
	t.Parallel()

	hook := &pipeForwardHook{
		writer: closedPipeWriter{},
		formatter: stubFormatter{
			output: []byte("ignored"),
		},
	}

	entry := newLogEntry(logrus.DebugLevel, "debug")

	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("expected nil error for closed pipe, got %v", err)
	}
}

func TestPipeForwardHookPropagatesFormatterErrors(t *testing.T) {
	t.Parallel()

	hook := &pipeForwardHook{
		writer:    &bytes.Buffer{},
		formatter: stubFormatter{err: errFormatFailed},
	}

	entry := newLogEntry(logrus.InfoLevel, "fail")

	err := hook.Fire(entry)
	if err == nil || !errors.Is(err, errFormatFailed) {
		t.Fatalf("expected formatter error, got %v", err)
	}
}

func TestPipeForwardHookReturnsWriteErrors(t *testing.T) {
	t.Parallel()

	hook := &pipeForwardHook{
		writer:    errorWriter{err: errWriteFailed},
		formatter: stubFormatter{output: []byte("data")},
	}

	entry := newLogEntry(logrus.InfoLevel, "warn")

	err := hook.Fire(entry)
	if err == nil || !errors.Is(err, errWriteFailed) {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestPipeForwardHookSkipsWhenUnconfigured(t *testing.T) {
	t.Parallel()

	cases := map[string]pipeForwardHook{
		"missing-writer":    {formatter: stubFormatter{output: []byte("noop")}},
		"missing-formatter": {writer: &bytes.Buffer{}},
	}

	entry := newLogEntry(logrus.InfoLevel, "skip")

	for name, hook := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := hook.Fire(entry)
			if err != nil {
				t.Fatalf("expected nil error for %s, got %v", name, err)
			}
		})
	}
}

func TestCommandExecutionBuildDestWriter(t *testing.T) {
	t.Parallel()

	t.Run("returns buffer when original nil", func(t *testing.T) {
		t.Parallel()

		runBuildDestWriterBufferCase(t)
	})

	t.Run("multi-writer mirrors output", func(t *testing.T) {
		t.Parallel()

		runBuildDestWriterMultiWriterCase(t)
	})
}

func runBuildDestWriterBufferCase(t *testing.T) {
	t.Helper()

	var buf bytes.Buffer

	ce := &commandExecution{}
	writer := ce.buildDestWriter(nil, &buf)

	_, err := writer.Write([]byte("payload"))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if buf.String() != "payload" {
		t.Fatalf("expected buffer to capture payload, got %q", buf.String())
	}
}

func runBuildDestWriterMultiWriterCase(t *testing.T) {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "commandrunner-dest-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	t.Cleanup(func() {
		_ = file.Close()
	})

	var buf bytes.Buffer

	ce := &commandExecution{}
	writer := ce.buildDestWriter(file, &buf)
	data := []byte("mirrored output")

	_, err = writer.Write(data)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if buf.String() != string(data) {
		t.Fatalf("buffer mismatch: got %q", buf.String())
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("failed to seek file: %v", err)
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(fileData) != string(data) {
		t.Fatalf("file mismatch: got %q", string(fileData))
	}
}

func TestCommandExecutionConfigureLoggerNil(t *testing.T) {
	t.Parallel()

	ce := &commandExecution{}

	err := ce.configureLogger()
	if !errors.Is(err, errLoggerUnavailable) {
		t.Fatalf("expected errLoggerUnavailable, got %v", err)
	}
}

func TestStripStdoutInfoHooks(t *testing.T) {
	t.Parallel()

	t.Run("removes info hooks targeting stdout", func(t *testing.T) {
		t.Parallel()

		stdout := newStdoutFile(t)

		hooks := logrus.LevelHooks{
			logrus.InfoLevel: []logrus.Hook{
				&logwriter.Hook{
					Writer:    stdout,
					LogLevels: []logrus.Level{logrus.InfoLevel},
				},
			},
			logrus.ErrorLevel: []logrus.Hook{
				&logwriter.Hook{
					Writer:    io.Discard,
					LogLevels: []logrus.Level{logrus.ErrorLevel},
				},
			},
		}

		filtered := stripStdoutInfoHooks(hooks, stdout)

		if len(filtered[logrus.InfoLevel]) != 0 {
			t.Fatalf("expected info hooks removed, got %d", len(filtered[logrus.InfoLevel]))
		}

		if len(filtered[logrus.ErrorLevel]) != 1 {
			t.Fatalf("expected error hook preserved")
		}
	})

	t.Run("returns nil for nil hooks", func(t *testing.T) {
		t.Parallel()

		stdout := newStdoutFile(t)

		if result := stripStdoutInfoHooks(nil, stdout); result != nil {
			t.Fatalf("expected nil, got %#v", result)
		}
	})
}

func TestIsStdoutInfoWriterHook(t *testing.T) {
	t.Parallel()

	stdout := newStdoutFile(t)

	stdoutHook := &logwriter.Hook{Writer: stdout, LogLevels: []logrus.Level{logrus.InfoLevel}}
	warningHook := &logwriter.Hook{Writer: stdout, LogLevels: []logrus.Level{logrus.WarnLevel}}
	otherWriterHook := &logwriter.Hook{
		Writer:    io.Discard,
		LogLevels: []logrus.Level{logrus.InfoLevel},
	}

	if !isStdoutInfoWriterHook(stdoutHook, stdout) {
		t.Fatal("expected stdout info hook to be detected")
	}

	if isStdoutInfoWriterHook(warningHook, stdout) {
		t.Fatal("did not expect warn-only hook to match")
	}

	if isStdoutInfoWriterHook(otherWriterHook, stdout) {
		t.Fatal("did not expect hook with different writer to match")
	}

	if isStdoutInfoWriterHook(stdoutHook, nil) {
		t.Fatal("did not expect match when stdout is nil")
	}
}

func TestCloneHooks(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when input nil", func(t *testing.T) {
		t.Parallel()

		if clone := cloneHooks(nil); clone != nil {
			t.Fatalf("expected nil clone, got %#v", clone)
		}
	})

	t.Run("skips empty levels and clones hooks", func(t *testing.T) {
		t.Parallel()

		stdout := newStdoutFile(t)

		original := logrus.LevelHooks{
			logrus.InfoLevel: {},
			logrus.ErrorLevel: {
				&logwriter.Hook{Writer: stdout, LogLevels: []logrus.Level{logrus.ErrorLevel}},
			},
		}

		clone := cloneHooks(original)

		if _, exists := clone[logrus.InfoLevel]; exists {
			t.Fatal("expected empty level to be skipped in clone")
		}

		actual := len(clone[logrus.ErrorLevel])
		expected := len(original[logrus.ErrorLevel])

		if actual != expected {
			t.Fatalf("expected error level hooks copied, got %d vs %d", actual, expected)
		}

		clone[logrus.ErrorLevel] = nil
		if original[logrus.ErrorLevel] == nil {
			t.Fatal("expected original hooks to remain unchanged")
		}
	})
}

type stubFormatter struct {
	output []byte
	err    error
}

func (s stubFormatter) Format(*logrus.Entry) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.output, nil
}

type closedPipeWriter struct{}

func (closedPipeWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func newLogEntry(level logrus.Level, message string) *logrus.Entry {
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	entry.Level = level
	entry.Message = message

	return entry
}

func newStdoutFile(t *testing.T) *os.File {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "commandrunner-stdout-*.log")
	if err != nil {
		t.Fatalf("failed to create temp stdout: %v", err)
	}

	t.Cleanup(func() {
		_ = file.Close()
	})

	return file
}

func setupCommandExecution(t *testing.T) *commandExecution {
	t.Helper()

	execCtx, err := newCommandExecution()
	if err != nil {
		t.Fatalf("new command execution: %v", err)
	}

	t.Cleanup(func() {
		execCtx.resetLogger()
		execCtx.restore()
	})

	return execCtx
}
