package notify_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	notify "github.com/devantler-tech/ksail-go/cmd/ui/notify"
)

// writer helpers no longer needed; we call *To(out, ...) variants directly

func TestErrorf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Errorf(&out, "%s: %d", "oops", 42)
	got := out.String()
	want := notify.ErrorSymbol + "oops: 42\n"

	// Assert
	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestError(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Error(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops"

	// Assert
	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestErrorln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Errorln(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops\n"

	// Assert
	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestWarnf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Warnf(&out, "%s", "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarn(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Warn(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarnln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Warnln(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Successf(&out, "%s", "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccess(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Success(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Successln(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Activityf(&out, "%s", "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivity(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Activity(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Activityln(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

// errorWriter is a mock writer that always returns an error.
type errorWriter struct{}

var errMockWrite = errors.New("mock write error")

func (ew errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("%w", errMockWrite)
}

func TestHandleNotifyError_WithError(t *testing.T) {
	t.Parallel()

	// Arrange
	// We'll capture stderr to verify the error handling
	oldStderr := os.Stderr
	readPipe, writePipe, _ := os.Pipe()
	os.Stderr = writePipe

	// Use an errorWriter to trigger the error path in handleNotifyError
	errWriter := errorWriter{}

	// Act
	notify.Error(errWriter, "test message")

	// Restore stderr
	err := writePipe.Close()
	if err != nil {
		t.Fatalf("failed to close writePipe: %v", err)
	}

	os.Stderr = oldStderr

	// Read what was written to stderr
	buf := make([]byte, 1024)
	n, _ := readPipe.Read(buf)
	output := string(buf[:n])

	// Assert
	expectedErrorMsg := "notify: failed to print message: mock write error\n"
	if output != expectedErrorMsg {
		t.Fatalf("expected stderr output %q, got %q", expectedErrorMsg, output)
	}
}
