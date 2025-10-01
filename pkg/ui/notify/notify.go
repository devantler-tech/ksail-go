// Package notify provides utilities for sending notifications to the user.
package notify

import (
	"fmt"
	"io"
	"os"
	"time"

	fcolor "github.com/fatih/color"
)

const (
	// ErrorSymbol is the symbol used for error messages.
	ErrorSymbol = "✗"
	// WarningSymbol is the symbol used for warning messages.
	WarningSymbol = "⚠"
	// SuccessSymbol is the symbol used for success messages.
	SuccessSymbol = "✔"
	// ActivitySymbol is the symbol used for activity messages.
	ActivitySymbol = "►"
	// InfoSymbol is the symbol used for informational messages.
	InfoSymbol = "ℹ"
)

// Message represents a notification message with optional timing information.
type Message struct {
	Text    string
	Elapsed time.Duration
	Stage   time.Duration
}

// NewMessage creates a new Message with the given text.
func NewMessage(text string) Message {
	return Message{Text: text}
}

// WithElapsed sets the elapsed duration for the message.
func (m Message) WithElapsed(d time.Duration) Message {
	m.Elapsed = d

	return m
}

// WithStage sets the stage duration for the message.
func (m Message) WithStage(d time.Duration) Message {
	m.Stage = d

	return m
}

// WithTiming sets both elapsed and stage durations for the message.
func (m Message) WithTiming(elapsed, stage time.Duration) Message {
	m.Elapsed = elapsed
	m.Stage = stage

	return m
}

// Format returns the formatted message text with timing if present.
func (m Message) Format() string {
	msg := m.Text
	if m.Elapsed > 0 && m.Stage > 0 {
		msg += fmt.Sprintf(" [%s|%s]",
			FormatDuration(m.Elapsed),
			FormatDuration(m.Stage))
	}

	return msg
}

// FormatDuration formats a duration as "2s", "45s", "1m30s", etc.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}

	return d.Truncate(time.Second).String()
}

// TitleMessage prints a title message (emoji + Message) with a trailing newline.
func TitleMessage(out io.Writer, emoji string, msg Message) {
	color := fcolor.New(fcolor.Reset, fcolor.Bold)
	writeMessage(out, color, emoji, msg.Format(), true)
}

// SuccessMessage prints a green success message with optional timing information.
func SuccessMessage(out io.Writer, msg Message) {
	color := fcolor.New(fcolor.FgGreen)
	writeMessage(out, color, SuccessSymbol, msg.Format(), true)
}

// ActivityMessage prints an activity message with optional timing information.
func ActivityMessage(out io.Writer, msg Message) {
	color := fcolor.New(fcolor.Reset)
	writeMessage(out, color, ActivitySymbol, msg.Format(), true)
}

// ErrorMessage prints an error message with optional timing information.
func ErrorMessage(out io.Writer, msg Message) {
	color := fcolor.New(fcolor.FgRed)
	writeMessage(out, color, ErrorSymbol, msg.Format(), true)
}

// InfoMessage prints an info message with optional timing information.
func InfoMessage(out io.Writer, msg Message) {
	color := fcolor.New(fcolor.FgBlue)
	writeMessage(out, color, InfoSymbol, msg.Format(), true)
}

// WarnMessage prints a warning message with optional timing information.
func WarnMessage(out io.Writer, msg Message) {
	color := fcolor.New(fcolor.FgYellow)
	writeMessage(out, color, WarningSymbol, msg.Format(), true)
}

// --- internals ---

// writeMessage prints a message with optional newline inserting a space between symbol
// and message when message present.
func writeMessage(
	out io.Writer,
	col *fcolor.Color,
	symbol, message string,
	newline bool,
) {
	content := symbol
	if message != "" {
		content += " " + message
	}

	if newline {
		content += "\n"
	}

	_, err := col.Fprint(out, content)
	handleNotifyError(err)
}

// handleNotifyError handles errors that occur during notification printing.
func handleNotifyError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "notify: failed to print message: %v\n", err)
	}
}
