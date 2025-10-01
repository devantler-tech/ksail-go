package notify_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	notify "github.com/devantler-tech/ksail-go/pkg/ui/notify"
)

// TestMessage tests the Message struct constructor and methods.
func TestMessage(t *testing.T) {
	t.Parallel()

	t.Run("NewMessage creates message with text only", func(t *testing.T) {
		t.Parallel()

		msg := notify.NewMessage("test message")
		if msg.Text != "test message" {
			t.Fatalf("expected text 'test message', got %q", msg.Text)
		}

		if msg.Elapsed != 0 {
			t.Fatalf("expected elapsed to be 0, got %v", msg.Elapsed)
		}

		if msg.Stage != 0 {
			t.Fatalf("expected stage to be 0, got %v", msg.Stage)
		}
	})

	t.Run("WithElapsed sets elapsed time", func(t *testing.T) {
		t.Parallel()

		duration := 5 * time.Second
		msg := notify.NewMessage("test").WithElapsed(duration)

		if msg.Elapsed != duration {
			t.Fatalf("expected elapsed %v, got %v", duration, msg.Elapsed)
		}
	})

	t.Run("WithStage sets stage time", func(t *testing.T) {
		t.Parallel()

		duration := 2 * time.Second
		msg := notify.NewMessage("test").WithStage(duration)

		if msg.Stage != duration {
			t.Fatalf("expected stage %v, got %v", duration, msg.Stage)
		}
	})

	t.Run("WithTiming sets both elapsed and stage", func(t *testing.T) {
		t.Parallel()

		elapsed := 10 * time.Second
		stage := 3 * time.Second
		msg := notify.NewMessage("test").WithTiming(elapsed, stage)

		if msg.Elapsed != elapsed {
			t.Fatalf("expected elapsed %v, got %v", elapsed, msg.Elapsed)
		}

		if msg.Stage != stage {
			t.Fatalf("expected stage %v, got %v", stage, msg.Stage)
		}
	})

	t.Run("Format returns text only when no timing", func(t *testing.T) {
		t.Parallel()

		msg := notify.NewMessage("simple message")
		formatted := msg.Format()

		if formatted != "simple message" {
			t.Fatalf("expected 'simple message', got %q", formatted)
		}
	})

	t.Run("Format returns text only when only elapsed is set", func(t *testing.T) {
		t.Parallel()

		msg := notify.NewMessage("with timing").WithElapsed(5 * time.Second)
		formatted := msg.Format()
		expected := "with timing"

		if formatted != expected {
			t.Fatalf("expected %q, got %q", expected, formatted)
		}
	})

	t.Run("Format includes timing when both are set", func(t *testing.T) {
		t.Parallel()

		msg := notify.NewMessage("full timing").WithTiming(10*time.Second, 3*time.Second)
		formatted := msg.Format()
		expected := "full timing [10s|3s]"

		if formatted != expected {
			t.Fatalf("expected %q, got %q", expected, formatted)
		}
	})
}

// TestFormatDuration tests duration formatting.
func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero duration", 0, "0s"},
		{"milliseconds under 1s", 500 * time.Millisecond, "0s"},
		{"seconds", 5 * time.Second, "5s"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2m30s"},
		{"hours minutes seconds", 1*time.Hour + 15*time.Minute + 45*time.Second, "1h15m45s"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := notify.FormatDuration(testCase.duration)
			if got != testCase.want {
				t.Fatalf("FormatDuration(%v) = %q, want %q", testCase.duration, got, testCase.want)
			}
		})
	}
}

// TestErrorMessage tests error message printing.
func TestErrorMessage(t *testing.T) {
	t.Parallel()

	t.Run("prints simple error message", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.ErrorMessage(&out, notify.NewMessage("oops"))
		got := out.String()
		want := notify.ErrorSymbol + " oops\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("prints error message without timing when only elapsed is set", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		msg := notify.NewMessage("failed").WithElapsed(2 * time.Second)
		notify.ErrorMessage(&out, msg)
		got := out.String()
		want := notify.ErrorSymbol + " failed\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// TestWarnMessage tests warning message printing.
func TestWarnMessage(t *testing.T) {
	t.Parallel()

	t.Run("prints simple warning", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.WarnMessage(&out, notify.NewMessage("careful"))
		got := out.String()
		want := notify.WarningSymbol + " careful\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("prints warning with timing", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		msg := notify.NewMessage("slow process").WithTiming(30*time.Second, 5*time.Second)
		notify.WarnMessage(&out, msg)
		got := out.String()
		want := notify.WarningSymbol + " slow process [30s|5s]\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// TestSuccessMessage tests success message printing.
func TestSuccessMessage(t *testing.T) {
	t.Parallel()

	t.Run("prints simple success", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.SuccessMessage(&out, notify.NewMessage("done"))
		got := out.String()
		want := notify.SuccessSymbol + " done\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("prints success without timing when only elapsed is set", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		msg := notify.NewMessage("completed").WithElapsed(10 * time.Second)
		notify.SuccessMessage(&out, msg)
		got := out.String()
		want := notify.SuccessSymbol + " completed\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// TestActivityMessage tests activity message printing.
func TestActivityMessage(t *testing.T) {
	t.Parallel()

	t.Run("prints simple activity", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.ActivityMessage(&out, notify.NewMessage("working"))
		got := out.String()
		want := notify.ActivitySymbol + " working\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("prints activity without timing when only stage is set", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		msg := notify.NewMessage("processing").WithStage(3 * time.Second)
		notify.ActivityMessage(&out, msg)
		got := out.String()
		want := notify.ActivitySymbol + " processing\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// TestInfoMessage tests info message printing.
func TestInfoMessage(t *testing.T) {
	t.Parallel()

	t.Run("prints simple info", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.InfoMessage(&out, notify.NewMessage("details"))
		got := out.String()
		want := notify.InfoSymbol + " details\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("prints info with full timing", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		msg := notify.NewMessage("metrics").WithTiming(1*time.Minute, 15*time.Second)
		notify.InfoMessage(&out, msg)
		got := out.String()
		want := notify.InfoSymbol + " metrics [1m0s|15s]\n"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// errorWriter is a mock writer that always returns an error.
type errorWriter struct{}

var errMockWrite = errors.New("mock write error")

func (ew errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("%w", errMockWrite)
}

func TestHandleNotifyErrorWithError(t *testing.T) {
	t.Parallel()

	// We'll capture stderr to verify the error handling
	oldStderr := os.Stderr
	readPipe, writePipe, _ := os.Pipe()
	os.Stderr = writePipe

	// Use an errorWriter to trigger the error path in handleNotifyError
	errWriter := errorWriter{}

	notify.ErrorMessage(errWriter, notify.NewMessage("test message"))

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

	expectedErrorMsg := "notify: failed to print message: mock write error\n"
	if output != expectedErrorMsg {
		t.Fatalf("expected stderr output %q, got %q", expectedErrorMsg, output)
	}
}

// Removed unused global buffer 'out'.

// TestTitleMessageBasic tests basic title message printing.
func TestTitleMessageBasic(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	notify.TitleMessage(&out, "✨", notify.NewMessage("Process finished successfully"))
	got := out.String()
	want := "✨ Process finished successfully\n"
	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestTitleMessageVariants(t *testing.T) {
	t.Parallel()

	t.Run("complex emoji", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.TitleMessage(&out, "🔧⚙️", notify.NewMessage("Configuration production loaded"))

		if out.String() != "🔧⚙️ Configuration production loaded\n" {
			t.Fatalf("unexpected output: %q", out.String())
		}
	})

	t.Run("unicode sequence", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.TitleMessage(&out, "📊", notify.NewMessage("Analytics dashboard ready"))

		if out.String() != "📊 Analytics dashboard ready\n" {
			t.Fatalf("unexpected output: %q", out.String())
		}
	})

	t.Run("empty emoji", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		notify.TitleMessage(&out, "", notify.NewMessage("No emoji title"))

		if out.String() != " No emoji title\n" {
			t.Fatalf("unexpected output: %q", out.String())
		}
	})
}
