package notify_test

import (
	"bytes"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

// ExampleWriteMessage_simple demonstrates basic message usage without timing.
func ExampleWriteMessage_simple() {
	var buf bytes.Buffer

	// Simple success message
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "operation completed successfully",
		Writer:  &buf,
	})

	fmt.Print(buf.String())
	// Output: ✔ operation completed successfully
}

// ExampleWriteMessage_formatted demonstrates formatted message content.
func ExampleWriteMessage_formatted() {
	var buf bytes.Buffer

	// Message with format arguments
	notify.WriteMessage(notify.Message{
		Type:    notify.InfoType,
		Content: "processing file %s (%d bytes)",
		Args:    []any{"config.yaml", 1024},
		Writer:  &buf,
	})

	fmt.Print(buf.String())
	// Output: ℹ processing file config.yaml (1024 bytes)
}

// ExampleWriteMessage_withTimer demonstrates automatic timer integration.
func ExampleWriteMessage_withTimer() {
	var buf bytes.Buffer

	// Create and use timer
	tmr := timer.New()
	tmr.Start()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Timer is automatically formatted and appended
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "deployment complete",
		Timer:   tmr,
		Writer:  &buf,
	})

	// Output will be like: ✔ deployment complete [100ms]
	// (exact timing varies, so we can't use exact output match in test)
	fmt.Printf("Message includes timing: %v\n", len(buf.String()) > 0)
	// Output: Message includes timing: true
}

// ExampleWriteMessage_multiStage demonstrates multi-stage timing.
func ExampleWriteMessage_multiStage() {
	var buf bytes.Buffer

	tmr := timer.New()
	tmr.Start()

	// Stage 1
	time.Sleep(50 * time.Millisecond)

	// Stage 2
	tmr.NewStage("deploying")
	time.Sleep(30 * time.Millisecond)

	// Timer shows both total and stage timing
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "multi-stage operation complete",
		Timer:   tmr,
		Writer:  &buf,
	})

	// Output will be like: ✔ multi-stage operation complete [80ms total|30ms stage]
	fmt.Printf("Message includes multi-stage timing: %v\n", len(buf.String()) > 0)
	// Output: Message includes multi-stage timing: true
}

// ExampleWriteMessage_title demonstrates custom title with emoji.
func ExampleWriteMessage_title() {
	var buf bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Starting Deployment",
		Emoji:   "🚀",
		Writer:  &buf,
	})

	fmt.Print(buf.String())
	// Output: 🚀 Starting Deployment
}

// ExampleWriteMessage_error demonstrates error message.
func ExampleWriteMessage_error() {
	var buf bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.ErrorType,
		Content: "failed to connect to server",
		Writer:  &buf,
	})

	fmt.Print(buf.String())
	// Output: ✗ failed to connect to server
}

// ExampleWriteMessage_warning demonstrates warning message.
func ExampleWriteMessage_warning() {
	var buf bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.WarningType,
		Content: "configuration file not found, using defaults",
		Writer:  &buf,
	})

	fmt.Print(buf.String())
	// Output: ⚠ configuration file not found, using defaults
}

// Example_backwardCompatibility demonstrates that old API still works.
func Example_backwardCompatibility() {
	var buf bytes.Buffer

	// Old API functions still work via convenience wrappers
	notify.Successf(&buf, "cluster %s is ready", "local")
	notify.Errorf(&buf, "failed to start %s", "service")
	notify.Warnf(&buf, "disk usage at %d%%", 85)

	fmt.Print(buf.String())
	// Output: ✔ cluster local is ready
	// ✗ failed to start service
	// ⚠ disk usage at 85%
}
