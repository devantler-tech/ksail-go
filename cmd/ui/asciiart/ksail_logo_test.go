package asciiart_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/asciiart"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPrintKSailLogo(t *testing.T) {
	// Arrange
	t.Parallel()

	var writer bytes.Buffer

	// Act
	asciiart.PrintKSailLogo(&writer)

	// Assert
	snaps.MatchSnapshot(t, writer.String())
}

// TestPrintKSailLogo_Comprehensive provides comprehensive testing of the public API
// to maximize code coverage achievable through the public interface only.
// Note: Due to //go:embed, some internal edge cases cannot be tested without
// modifying the source file and rebuilding (see scripts/test_edge_cases.sh).
func TestPrintKSailLogo_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "produces_non_empty_output",
			testFunc: func(t *testing.T) {
				var writer bytes.Buffer
				asciiart.PrintKSailLogo(&writer)
				output := writer.String()
				if len(output) == 0 {
					t.Error("Expected non-empty output")
				}
			},
		},
		{
			name: "contains_ascii_art_elements",
			testFunc: func(t *testing.T) {
				var writer bytes.Buffer
				asciiart.PrintKSailLogo(&writer)
				output := writer.String()

				// Verify presence of ASCII art characters
				expectedElements := []string{"__", "/", "\\", "|", "~", "^", "_", "-"}
				for _, element := range expectedElements {
					if !bytes.Contains([]byte(output), []byte(element)) {
						t.Errorf("Expected output to contain ASCII element %q", element)
					}
				}
			},
		},
		{
			name: "has_proper_line_structure",
			testFunc: func(t *testing.T) {
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
			},
		},
		{
			name: "output_is_consistent",
			testFunc: func(t *testing.T) {
				var writer1, writer2 bytes.Buffer

				asciiart.PrintKSailLogo(&writer1)
				asciiart.PrintKSailLogo(&writer2)

				if !bytes.Equal(writer1.Bytes(), writer2.Bytes()) {
					t.Error("Expected consistent output between multiple calls")
				}
			},
		},
		{
			name: "handles_different_writers",
			testFunc: func(t *testing.T) {
				writers := []bytes.Buffer{
					{}, // Fresh buffer
					{}, // Another fresh buffer
				}

				for i, writer := range writers {
					var w bytes.Buffer = writer
					asciiart.PrintKSailLogo(&w)
					output := w.String()
					if len(output) == 0 {
						t.Errorf("Writer %d produced empty output", i)
					}
				}
			},
		},
		{
			name: "validates_output_format",
			testFunc: func(t *testing.T) {
				var writer bytes.Buffer
				asciiart.PrintKSailLogo(&writer)
				output := writer.String()

				// Should end with newline
				if !bytes.HasSuffix([]byte(output), []byte("\n")) {
					t.Error("Expected output to end with newline")
				}

				// Should contain the KSail text pattern
				if !bytes.Contains([]byte(output), []byte("__")) {
					t.Error("Expected output to contain KSail ASCII pattern")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
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
		tc := tc
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
