package asciiart_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/asciiart"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPrintKSailLogo(t *testing.T) {
	t.Parallel()

	var writer bytes.Buffer

	asciiart.PrintKSailLogo(&writer)

	snaps.MatchSnapshot(t, writer.String())
}

// TestPrintKSailLogo_Comprehensive provides comprehensive testing of the public API
// to maximize code coverage achievable through the public interface only.
// Note: Due to //go:embed, some internal edge cases cannot be tested without
// modifying the source file and rebuilding (see scripts/test_edge_cases.sh).
func TestPrintKSailLogo_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("produces_non_empty_output", testNonEmptyOutput)
	t.Run("contains_ascii_art_elements", testASCIIElements)
	t.Run("has_proper_line_structure", testLineStructure)
	t.Run("output_is_consistent", testOutputConsistency)
	t.Run("handles_different_writers", testDifferentWriters)
	t.Run("validates_output_format", testOutputFormat)
}

func testNonEmptyOutput(t *testing.T) {
	t.Helper()
	t.Parallel()

	var writer bytes.Buffer
	asciiart.PrintKSailLogo(&writer)

	output := writer.String()
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

func testASCIIElements(t *testing.T) {
	t.Helper()
	t.Parallel()

	var writer bytes.Buffer
	asciiart.PrintKSailLogo(&writer)

	output := writer.String()

	// Verify presence of ASCII art characters
	expectedElements := []string{"__", "/", "\\", "|", "~", "^", "_", "-"}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain ASCII element %q", element)
		}
	}
}

func testLineStructure(t *testing.T) {
	t.Helper()
	t.Parallel()

	var writer bytes.Buffer
	asciiart.PrintKSailLogo(&writer)
	output := writer.String()

	// Verify multi-line structure
	lines := bytes.Split([]byte(output), []byte("\n"))
	if len(lines) < 8 {
		t.Errorf("Expected at least 8 lines of ASCII art, got %d", len(lines))
	}

	// Verify some lines have content (not all empty)
	nonEmptyLines := 0

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) > 0 {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 5 {
		t.Errorf("Expected at least 5 non-empty lines, got %d", nonEmptyLines)
	}
}

func testOutputConsistency(t *testing.T) {
	t.Helper()
	t.Parallel()

	var writer1, writer2 bytes.Buffer

	asciiart.PrintKSailLogo(&writer1)
	asciiart.PrintKSailLogo(&writer2)

	if !bytes.Equal(writer1.Bytes(), writer2.Bytes()) {
		t.Error("Expected consistent output between multiple calls")
	}
}

func testDifferentWriters(t *testing.T) {
	t.Helper()
	t.Parallel()

	writers := []bytes.Buffer{
		{}, // Fresh buffer
		{}, // Another fresh buffer
	}

	for index, writer := range writers {
		w := writer
		asciiart.PrintKSailLogo(&w)

		output := w.String()
		if len(output) == 0 {
			t.Errorf("Writer %d produced empty output", index)
		}
	}
}

func testOutputFormat(t *testing.T) {
	t.Helper()
	t.Parallel()

	var writer bytes.Buffer
	asciiart.PrintKSailLogo(&writer)
	output := writer.String()

	// Should end with newline
	if !strings.HasSuffix(output, "\n") {
		t.Error("Expected output to end with newline")
	}

	// Should contain the KSail text pattern
	if !strings.Contains(output, "__") {
		t.Error("Expected output to contain KSail ASCII pattern")
	}
}

// TestPrintKSailLogo_Writers tests the function with various io.Writer implementations
// to ensure the public API works correctly with different output destinations.
func TestPrintKSailLogo_Writers(t *testing.T) {
	t.Parallel()

	// Test with different writer scenarios
	testCases := []struct {
		name   string
		writer func() *bytes.Buffer
	}{
		{
			name: "fresh_buffer",
			writer: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
		{
			name: "buffer_with_existing_content",
			writer: func() *bytes.Buffer {
				buf := &bytes.Buffer{}
				buf.WriteString("prefix_")

				return buf
			},
		},
		{
			name: "large_capacity_buffer",
			writer: func() *bytes.Buffer {
				buf := &bytes.Buffer{}
				buf.Grow(1000) // Pre-allocate space

				return buf
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			writer := tc.writer()
			initialLen := writer.Len()

			asciiart.PrintKSailLogo(writer)

			// Verify content was written
			if writer.Len() <= initialLen {
				t.Error("Expected writer to have more content after PrintKSailLogo")
			}

			output := writer.String()

			// Extract just the logo part (after any prefix)
			logoOutput := output[initialLen:]
			if len(logoOutput) == 0 {
				t.Error("Expected logo output to be written to buffer")
			}
		})
	}
}

// TestPrintKSailLogo_BoundsHandling tests that the function handles lines
// of various lengths without panicking, demonstrating the robustness of
// the bounds checking implementation.
func TestPrintKSailLogo_BoundsHandling(t *testing.T) {
	t.Parallel()

	var writer bytes.Buffer

	// This test verifies that the function doesn't panic on the actual logo content
	// which includes lines of different lengths that could trigger bounds issues
	// if proper length checking wasn't implemented.
	asciiart.PrintKSailLogo(&writer)

	output := writer.String()

	// Verify we got output without panics
	if len(output) == 0 {
		t.Error("Expected non-empty output from logo printing")
	}

	// Verify the output contains expected content
	if !strings.Contains(output, "KSail") && !strings.Contains(output, "__") {
		t.Error("Expected output to contain recognizable logo elements")
	}

	// Verify proper line structure (no panics means bounds checking worked)
	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("Expected multiple lines in output, got %d", len(lines))
	}
}
