// Package notify provides utilities for sending notifications to the user.
package notify

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	fcolor "github.com/fatih/color"
)

// MessageType defines the type of notification message.
type MessageType int

const (
	// ErrorType represents an error message (red, with ✗ symbol).
	ErrorType MessageType = iota
	// WarningType represents a warning message (yellow, with ⚠ symbol).
	WarningType
	// ActivityType represents an activity/progress message (default color, with ► symbol).
	ActivityType
	// SuccessType represents a success message (green, with ✔ symbol).
	SuccessType
	// InfoType represents an informational message (blue, with ℹ symbol).
	InfoType
	// TitleType represents a title/header message (bold, with emoji (custom or default)).
	TitleType
)

// Message represents a notification message to be displayed to the user.
type Message struct {
	// Type determines the message styling (color, symbol).
	Type MessageType
	// Content is the main message text to display.
	Content string
	// Timer is optional. If provided, timing information will be appended to the message.
	Timer timer.Timer
	// Emoji is used only for TitleType messages to customize the title icon.
	Emoji string
	// Writer is the output destination. If nil, defaults to os.Stdout.
	Writer io.Writer
	// Args are format arguments for Content if it contains format specifiers.
	Args []any
}

// WriteMessage writes a formatted message based on the message configuration.
// It handles message styling, optional timing information, and proper output formatting.
func WriteMessage(msg Message) {
	// Default to stdout if no writer specified
	if msg.Writer == nil {
		msg.Writer = os.Stdout
	}

	// Format the message content
	content := msg.Content
	if len(msg.Args) > 0 {
		content = fmt.Sprintf(msg.Content, msg.Args...)
	}

	// Append timing information if timer is provided
	if msg.Timer != nil {
		total, stage := msg.Timer.GetTiming()
		timingStr := FormatTiming(total, stage, total != stage)
		content = fmt.Sprintf("%s %s", content, timingStr)
	}

	// Get message configuration based on type
	config := getMessageConfig(msg.Type)

	// Handle TitleType specially (uses emoji instead of symbol)
	if msg.Type == TitleType {
		emoji := msg.Emoji
		if emoji == "" {
			emoji = "ℹ️" // default emoji for titles
		}

		_, err := config.color.Fprintf(msg.Writer, "%s %s\n", emoji, content)
		handleNotifyError(err)

		return
	}

	// Write message with symbol and color
	_, err := config.color.Fprintf(msg.Writer, "%s%s\n", config.symbol, content)
	handleNotifyError(err)
}

// messageConfig holds the styling configuration for each message type.
type messageConfig struct {
	symbol string
	color  *fcolor.Color
}

// getMessageConfig returns the styling configuration for a given message type.
func getMessageConfig(msgType MessageType) messageConfig {
	switch msgType {
	case ErrorType:
		return messageConfig{
			symbol: "✗ ",
			color:  fcolor.New(fcolor.FgRed),
		}
	case WarningType:
		return messageConfig{
			symbol: "⚠ ",
			color:  fcolor.New(fcolor.FgYellow),
		}
	case ActivityType:
		return messageConfig{
			symbol: "► ",
			color:  fcolor.New(fcolor.Reset),
		}
	case SuccessType:
		return messageConfig{
			symbol: "✔ ",
			color:  fcolor.New(fcolor.FgGreen),
		}
	case InfoType:
		return messageConfig{
			symbol: "ℹ ",
			color:  fcolor.New(fcolor.FgBlue),
		}
	case TitleType:
		return messageConfig{
			symbol: "",
			color:  fcolor.New(fcolor.Reset, fcolor.Bold),
		}
	default:
		return messageConfig{
			symbol: "",
			color:  fcolor.New(fcolor.Reset),
		}
	}
}

// handleNotifyError handles errors that occur during notification printing.
func handleNotifyError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "notify: failed to print message: %v\n", err)
	}
}

// FormatTiming formats timing durations into a display string.
// Returns "[stage: X|total: Y]" for multi-stage commands (when isMultiStage is true and total != stage)
// Returns "[stage: X]" for single-stage commands or when total == stage.
// Formats durations with a maximum of 2 decimal places.
func FormatTiming(total, stage time.Duration, isMultiStage bool) string {
	// If durations are equal or not multi-stage, use simplified format
	if !isMultiStage || total == stage {
		return fmt.Sprintf("[stage: %s]", formatDuration(total))
	}

	return fmt.Sprintf("[stage: %s|total: %s]", formatDuration(stage), formatDuration(total))
}

// formatDuration formats a duration with a maximum of 2 decimal places for sub-second durations.
// For durations >= 1 minute, uses compound format (e.g., "5m30s", "1h23m45s").
// For durations < 1 minute, shows with up to 2 decimals and appropriate unit.
// Trailing zeros are removed for cleaner output.
// Examples: 1.23s, 456.78ms, 12.34µs, 1.5s, 2s, 5m30s, 1h23m45s.
func formatDuration(duration time.Duration) string {
	// Handle special cases
	if duration == 0 {
		return "0s"
	}

	// For durations >= 1 minute, use compound format (hours, minutes, seconds)
	if duration >= time.Minute {
		return formatCompoundDuration(duration)
	}

	// For sub-minute durations, use decimal format with appropriate unit
	var (
		value float64
		unit  string
	)

	switch {
	case duration >= time.Second:
		value = float64(duration) / float64(time.Second)
		unit = "s"
	case duration >= time.Millisecond:
		value = float64(duration) / float64(time.Millisecond)
		unit = "ms"
	case duration >= time.Microsecond:
		value = float64(duration) / float64(time.Microsecond)
		unit = "µs"
	default:
		// For nanoseconds, just show the value without decimals
		return fmt.Sprintf("%dns", duration.Nanoseconds())
	}

	// Format with 2 decimal places and trim trailing zeros
	formatted := fmt.Sprintf("%.2f", value)
	// Remove trailing zeros and decimal point if not needed
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")

	return formatted + unit
}

// formatCompoundDuration formats durations >= 1 minute in compound format (e.g., "5m30s", "1h23m45s").
func formatCompoundDuration(duration time.Duration) string {
	hours := duration / time.Hour
	duration %= time.Hour

	minutes := duration / time.Minute
	duration %= time.Minute

	seconds := duration / time.Second

	// Build the string based on what components are present
	var parts []string

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	// Always show seconds if there are hours or minutes present
	// Only show seconds alone if it's the only component
	if seconds > 0 || len(parts) > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, "")
}
