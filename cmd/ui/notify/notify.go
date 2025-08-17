// Package notify provides utilities for sending notifications to the user.
package notify

import (
	"fmt"
	"io"
	"os"

	fcolor "github.com/fatih/color"
)

const (
	// ErrorSymbol is the symbol used for error messages.
	ErrorSymbol = "✗ "
	// WarningSymbol is the symbol used for warning messages.
	WarningSymbol = "⚠ "
	// SuccessSymbol is the symbol used for success messages.
	SuccessSymbol = "✔ "
	// ActivitySymbol is the symbol used for activity messages.
	ActivitySymbol = "► "
)

// Errorf prints a red error message to stderr, prefixed with a symbol.
func Errorf(format string, args ...any) { ErrorfTo(os.Stderr, format, args...) }

// ErrorfTo prints a red error message to the provided writer, prefixed with a symbol.
func ErrorfTo(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notifyf(out, color, ErrorSymbol, format, args...)
}

// Error prints a red error message to stderr without a trailing newline, prefixed with a symbol.
func Error(args ...any) { ErrorTo(os.Stderr, args...) }

// ErrorTo prints a red error message to the provided writer without a trailing newline, prefixed with a symbol.
func ErrorTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notify(out, color, ErrorSymbol, args...)
}

// Errorln prints a red error message to stderr with a trailing newline, prefixed with a symbol.
func Errorln(args ...any) { ErrorlnTo(os.Stderr, args...) }

// ErrorlnTo prints a red error message to the provided writer with a trailing newline, prefixed with a symbol.
func ErrorlnTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notifyln(out, color, ErrorSymbol, args...)
}

// Warnf prints a yellow warning message to stdout, prefixed with a symbol.
func Warnf(format string, args ...any) { WarnfTo(os.Stdout, format, args...) }

// WarnfTo prints a yellow warning message to the provided writer, prefixed with a symbol.
func WarnfTo(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notifyf(out, color, WarningSymbol, format, args...)
}

// Warn prints a yellow warning message to stdout without a trailing newline, prefixed with a symbol.
func Warn(args ...any) { WarnTo(os.Stdout, args...) }

// WarnTo prints a yellow warning message to the provided writer without a trailing newline, prefixed with a symbol.
func WarnTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notify(out, color, WarningSymbol, args...)
}

// Warnln prints a yellow warning message to stdout with a trailing newline, prefixed with a symbol.
func Warnln(args ...any) { WarnlnTo(os.Stdout, args...) }

// WarnlnTo prints a yellow warning message to the provided writer with a trailing newline, prefixed with a symbol.
func WarnlnTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notifyln(out, color, WarningSymbol, args...)
}

// Successf prints a green success message to stdout, prefixed with a symbol.
func Successf(format string, args ...any) { SuccessfTo(os.Stdout, format, args...) }

// SuccessfTo prints a green success message to the provided writer, prefixed with a symbol.
func SuccessfTo(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notifyf(out, color, SuccessSymbol, format, args...)
}

// Success prints a green success message to stdout without a trailing newline, prefixed with a symbol.
func Success(args ...any) { SuccessTo(os.Stdout, args...) }

// SuccessTo prints a green success message to the provided writer without a trailing newline, prefixed with a symbol.
func SuccessTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notify(out, color, SuccessSymbol, args...)
}

// Successln prints a green success message to stdout with a trailing newline, prefixed with a symbol.
func Successln(args ...any) { SuccesslnTo(os.Stdout, args...) }

// SuccesslnTo prints a green success message to the provided writer with a trailing newline, prefixed with a symbol.
func SuccesslnTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notifyln(out, color, SuccessSymbol, args...)
}

// Activityf prints a blue activity message to stdout, prefixed with a symbol.
func Activityf(format string, args ...any) { ActivityfTo(os.Stdout, format, args...) }

// ActivityfTo prints a blue activity message to the provided writer, prefixed with a symbol.
func ActivityfTo(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgBlue)
	notifyf(out, color, ActivitySymbol, format, args...)
}

// Activity prints a blue activity message to stdout without a trailing newline, prefixed with a symbol.
func Activity(args ...any) { ActivityTo(os.Stdout, args...) }

// ActivityTo prints a blue activity message to the provided writer without a trailing newline, prefixed with a symbol.
func ActivityTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgBlue)
	notify(out, color, ActivitySymbol, args...)
}

// Activityln prints a blue activity message to stdout with a trailing newline, prefixed with a symbol.
func Activityln(args ...any) { ActivitylnTo(os.Stdout, args...) }

// ActivitylnTo prints a blue activity message to the provided writer with a trailing newline, prefixed with a symbol.
func ActivitylnTo(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgBlue)
	notifyln(out, color, ActivitySymbol, args...)
}

// --- internals ---

// notifyf prints a symbol and a formatted message with a trailing newline using the provided color and writer.
func notifyf(out io.Writer, col *fcolor.Color, symbol, format string, args ...any) {
	allArgs := append([]any{symbol}, args...)
	_, err := col.Fprintf(out, "%s"+format+"\n", allArgs...)
	handleNotifyError(err)
}

// notify prints a symbol and message without a trailing newline using the provided color and writer.
func notify(out io.Writer, col *fcolor.Color, symbol string, args ...any) {
	_, err := col.Fprint(out, symbol, fmt.Sprint(args...))
	handleNotifyError(err)
}

// notifyln prints a symbol and message with a trailing newline using the provided color and writer.
func notifyln(out io.Writer, col *fcolor.Color, symbol string, args ...any) {
	_, err := col.Fprintln(out, symbol, fmt.Sprint(args...))
	handleNotifyError(err)
}

// notifyf prints a symbol and a formatted message with a trailing newline using the provided color and writer.
func handleNotifyError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "notify: failed to print message: %v\n", err)
	}
}
