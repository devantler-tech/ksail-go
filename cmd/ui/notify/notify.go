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
func Errorf(format string, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notifyf(os.Stderr, color, ErrorSymbol, format, args...)
}

// Error prints a red error message to stderr without a trailing newline, prefixed with a symbol.
func Error(args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notify(os.Stderr, color, ErrorSymbol, args...)
}

// Errorln prints a red error message to stderr with a trailing newline, prefixed with a symbol.
func Errorln(args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notifyln(os.Stderr, color, ErrorSymbol, args...)
}

// Warnf prints a yellow warning message to stdout, prefixed with a symbol.
func Warnf(format string, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notifyf(os.Stdout, color, WarningSymbol, format, args...)
}

// Warn prints a yellow warning message to stdout without a trailing newline, prefixed with a symbol.
func Warn(args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notify(os.Stdout, color, WarningSymbol, args...)
}

// Warnln prints a yellow warning message to stdout with a trailing newline, prefixed with a symbol.
func Warnln(args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notifyln(os.Stdout, color, WarningSymbol, args...)
}

// Successf prints a green success message to stdout, prefixed with a symbol.
func Successf(format string, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notifyf(os.Stdout, color, SuccessSymbol, format, args...)
}

// Success prints a green success message to stdout without a trailing newline, prefixed with a symbol.
func Success(args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notify(os.Stdout, color, SuccessSymbol, args...)
}

// Successln prints a green success message to stdout with a trailing newline, prefixed with a symbol.
func Successln(args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notifyln(os.Stdout, color, SuccessSymbol, args...)
}

// Activityf prints a blue activity message to stdout, prefixed with a symbol.
func Activityf(format string, args ...any) {
	color := fcolor.New(fcolor.FgBlue)
	notifyf(os.Stdout, color, ActivitySymbol, format, args...)
}

// Activity prints a blue activity message to stdout without a trailing newline, prefixed with a symbol.
func Activity(args ...any) {
	color := fcolor.New(fcolor.FgBlue)
	notify(os.Stdout, color, ActivitySymbol, args...)
}

// Activityln prints a blue activity message to stdout with a trailing newline, prefixed with a symbol.
func Activityln(args ...any) {
	color := fcolor.New(fcolor.FgBlue)
	notifyln(os.Stdout, color, ActivitySymbol, args...)
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
