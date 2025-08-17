package notify_test

import (
	"bytes"
	"testing"

	notify "github.com/devantler-tech/ksail-go/cmd/ui/notify"
)

// writer helpers no longer needed; we call *To(out, ...) variants directly

func TestErrorf(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

  // Act
	notify.ErrorfTo(&out, "%s: %d", "oops", 42)
	got := out.String()
	want := notify.ErrorSymbol + "oops: 42\n"

  // Assert
	if got != want {
		test.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestError(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

  // Act
	notify.ErrorTo(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops"

	// Assert
	if got != want {
		test.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestErrorln(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.ErrorlnTo(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops\n"

	// Assert
	if got != want {
		test.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestWarnf(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.WarnfTo(&out, "%s", "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarn(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.WarnTo(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarnln(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.WarnlnTo(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessf(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

  // Act
	notify.SuccessfTo(&out, "%s", "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccess(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

  // Act
	notify.SuccessTo(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done"

  // Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessln(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.SuccesslnTo(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityf(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.ActivityfTo(&out, "%s", "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivity(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

  // Act
	notify.ActivityTo(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityln(test *testing.T) {
	// Arrange
	test.Parallel()

	var out bytes.Buffer

	// Act
	notify.ActivitylnTo(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	// Assert
	if got != want {
		test.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}
