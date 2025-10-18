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
	errBaseFailure  = errors.New("base failure")
	errFormatFailed = errors.New("format failed")
	errWriteFailed  = errors.New("write failed")
)

func TestCobraCommandRunner_RunPropagatesStdout(t *testing.T) {
	t.Parallel()

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

		hooks := logrus.LevelHooks{
			logrus.InfoLevel: []logrus.Hook{
				&logwriter.Hook{
					Writer:    os.Stdout,
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

		filtered := stripStdoutInfoHooks(hooks, os.Stdout)

		if len(filtered[logrus.InfoLevel]) != 0 {
			t.Fatalf("expected info hooks removed, got %d", len(filtered[logrus.InfoLevel]))
		}

		if len(filtered[logrus.ErrorLevel]) != 1 {
			t.Fatalf("expected error hook preserved")
		}
	})

	t.Run("returns nil for nil hooks", func(t *testing.T) {
		t.Parallel()

		if result := stripStdoutInfoHooks(nil, os.Stdout); result != nil {
			t.Fatalf("expected nil, got %#v", result)
		}
	})
}

func TestIsStdoutInfoWriterHook(t *testing.T) {
	t.Parallel()

	stdoutHook := &logwriter.Hook{Writer: os.Stdout, LogLevels: []logrus.Level{logrus.InfoLevel}}
	warningHook := &logwriter.Hook{Writer: os.Stdout, LogLevels: []logrus.Level{logrus.WarnLevel}}
	otherWriterHook := &logwriter.Hook{
		Writer:    io.Discard,
		LogLevels: []logrus.Level{logrus.InfoLevel},
	}

	if !isStdoutInfoWriterHook(stdoutHook, os.Stdout) {
		t.Fatal("expected stdout info hook to be detected")
	}

	if isStdoutInfoWriterHook(warningHook, os.Stdout) {
		t.Fatal("did not expect warn-only hook to match")
	}

	if isStdoutInfoWriterHook(otherWriterHook, os.Stdout) {
		t.Fatal("did not expect hook with different writer to match")
	}

	if isStdoutInfoWriterHook(stdoutHook, nil) {
		t.Fatal("did not expect match when stdout is nil")
	}
}

func TestCloneHooks(t *testing.T) {
	t.Parallel()

	original := logrus.LevelHooks{
		logrus.InfoLevel: {
			&logwriter.Hook{Writer: os.Stdout, LogLevels: []logrus.Level{logrus.InfoLevel}},
		},
		logrus.ErrorLevel: {
			&logwriter.Hook{Writer: os.Stdout, LogLevels: []logrus.Level{logrus.ErrorLevel}},
		},
	}

	clone := cloneHooks(original)

	if len(clone) != len(original) {
		t.Fatalf("expected clone to match size, got %d vs %d", len(clone), len(original))
	}

	// Mutate clone and ensure original unaffected.
	clone[logrus.InfoLevel] = nil
	if original[logrus.InfoLevel] == nil {
		t.Fatal("expected original to remain unchanged")
	}
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
