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

// Errorf prints a red error message to the provided writer, prefixed with a symbol.
func Errorf(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notifyf(out, color, ErrorSymbol, format, args...)
}

// Error prints a red error message to the provided writer without a trailing newline, prefixed with a symbol.
func Error(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notify(out, color, ErrorSymbol, args...)
}

// Errorln prints a red error message to the provided writer with a trailing newline, prefixed with a symbol.
func Errorln(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgRed)
	notifyln(out, color, ErrorSymbol, args...)
}

// Warnf prints a yellow warning message to the provided writer, prefixed with a symbol.
func Warnf(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notifyf(out, color, WarningSymbol, format, args...)
}

// Warn prints a yellow warning message to the provided writer without a trailing newline, prefixed with a symbol.
func Warn(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notify(out, color, WarningSymbol, args...)
}

// Warnln prints a yellow warning message to the provided writer with a trailing newline, prefixed with a symbol.
func Warnln(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgYellow)
	notifyln(out, color, WarningSymbol, args...)
}

// Successf prints a green success message to the provided writer, prefixed with a symbol.
func Successf(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notifyf(out, color, SuccessSymbol, format, args...)
}

// Success prints a green success message to the provided writer without a trailing newline, prefixed with a symbol.
func Success(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notify(out, color, SuccessSymbol, args...)
}

// Successln prints a green success message to the provided writer with a trailing newline, prefixed with a symbol.
func Successln(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.FgGreen)
	notifyln(out, color, SuccessSymbol, args...)
}

// Activityf prints a blue activity message to the provided writer, prefixed with a symbol.
func Activityf(out io.Writer, format string, args ...any) {
	color := fcolor.New(fcolor.Reset)
	notifyf(out, color, ActivitySymbol, format, args...)
}

// Activity prints a blue activity message to the provided writer without a trailing newline, prefixed with a symbol.
func Activity(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.Reset)
	notify(out, color, ActivitySymbol, args...)
}

// Activityln prints a blue activity message to the provided writer with a trailing newline, prefixed with a symbol.
func Activityln(out io.Writer, args ...any) {
	color := fcolor.New(fcolor.Reset)
	notifyln(out, color, ActivitySymbol, args...)
}

// Titlef prints a formatted title message to the provided writer with an emoji and title.
func Titlef(out io.Writer, emoji, format string, args ...any) {
	color := fcolor.New(fcolor.Reset, fcolor.Bold)
	titlef(out, color, emoji, format, args...)
}

// Title prints a title message to the provided writer with an emoji and title.
func Title(out io.Writer, emoji string, args ...any) {
	color := fcolor.New(fcolor.Reset, fcolor.Bold)
	title(out, color, emoji, args...)
}

// Titleln prints a title message to the provided writer with a trailing newline, using an emoji and title.
func Titleln(out io.Writer, emoji string, args ...any) {
	color := fcolor.New(fcolor.Reset, fcolor.Bold)
	titleln(out, color, emoji, args...)
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
	_, err := col.Fprintln(out, symbol+fmt.Sprint(args...))
	handleNotifyError(err)
}

// titlef prints an emoji and a formatted title message with a trailing newline using the provided color and writer.
func titlef(out io.Writer, col *fcolor.Color, emoji, format string, args ...any) {
	titleText := fmt.Sprintf(format, args...)
	_, err := col.Fprintf(out, "%s %s\n", emoji, titleText)
	handleNotifyError(err)
}

// title prints an emoji and title message without a trailing newline using the provided color and writer.
func title(out io.Writer, col *fcolor.Color, emoji string, args ...any) {
	titleText := fmt.Sprint(args...)
	_, err := col.Fprint(out, emoji, " ", titleText)
	handleNotifyError(err)
}

// titleln prints an emoji and title message with a trailing newline using the provided color and writer.
func titleln(out io.Writer, col *fcolor.Color, emoji string, args ...any) {
	titleText := fmt.Sprint(args...)
	_, err := col.Fprintln(out, emoji, titleText)
	handleNotifyError(err)
}

// handleNotifyError handles errors that occur during notification printing.
func handleNotifyError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "notify: failed to print message: %v\n", err)
	}
}
