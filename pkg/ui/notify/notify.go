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

// Message type constants.
// Each type determines the message styling (color and symbol).
const (
	// ErrorType represents an error message (red, with ✗ symbol).
	ErrorType MessageType = iota
	// WarningType represents a warning message (yellow, with ⚠ symbol).
	WarningType
	// ActivityType represents an activity/progress message (default color, with ► symbol).
	ActivityType
	// GenerateType represents a file generation message (default color, with ✚ symbol).
	GenerateType
	// SuccessType represents a success message (green, with ✔ symbol).
	SuccessType
	// InfoType represents an informational message (blue, with ℹ symbol).
	InfoType
	// TitleType represents a title/header message (bold, with emoji (custom or default)).
	TitleType
)

// MessageType defines the type of notification message.
type MessageType int

// Message represents a notification message to be displayed to the user.
type Message struct {
	// Type determines the message styling (color, symbol).
	Type MessageType
	// Content is the main message text to display.
	Content string
	// Timer is optional. If provided, timing information will be appended to the message.
	Timer timer.Timer
	// MultiStage MUST be set to true for multi-stage timers (i.e., when the command advances through stages).
	// If false, the timing output will be rendered in single-stage form regardless of internal durations.
	MultiStage bool
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
		// Use explicit MultiStage flag only; heuristic removed to avoid accidental misclassification.
		timingStr := FormatTiming(total, stage, msg.MultiStage)
		content = fmt.Sprintf("%s %s", content, timingStr)
	}

	// Get message configuration based on type
	config := getMessageConfig(msg.Type)

	content = indentMultilineContent(content, config.symbol)

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

// Message configuration helpers.

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
	case GenerateType:
		return messageConfig{
			symbol: "✚ ",
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

// Error handling helpers.

// handleNotifyError handles errors that occur during notification printing.
// Errors are logged to stderr rather than returned to avoid disrupting the user experience.
func handleNotifyError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "notify: failed to print message: %v\n", err)
	}
}

// Timing formatting helpers.

// FormatTiming formats timing durations into a display string using Go's Duration.String() method.
// Returns "[stage: X|total: Y]" for multi-stage commands when isMultiStage is true.
// Returns "[stage: X]" for single-stage commands.
// Uses Go's standard Duration.String() which provides appropriate precision automatically.
func FormatTiming(total, stage time.Duration, isMultiStage bool) string {
	if !isMultiStage {
		return fmt.Sprintf("[stage: %s]", stage.String())
	}

	return fmt.Sprintf("[stage: %s|total: %s]", stage.String(), total.String())
}

// Content formatting helpers.

// indentMultilineContent indents subsequent lines of multi-line content based on the symbol width.
// This ensures that multi-line messages are properly aligned with the first line's symbol.
func indentMultilineContent(content, symbol string) string {
	if symbol == "" || !strings.Contains(content, "\n") {
		return content
	}

	indent := strings.Repeat(" ", len([]rune(symbol)))
	lines := strings.Split(content, "\n")

	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}

		lines[i] = indent + lines[i]
	}

	return strings.Join(lines, "\n")
}
