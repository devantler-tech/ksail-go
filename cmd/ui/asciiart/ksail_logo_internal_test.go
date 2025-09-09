package asciiart

import (
	"bytes"
	"strings"
	"testing"
)

// TestPrintGreenBlueCyanPart_ShortLine tests the edge case where the line
// is shorter than cyanStartIndex, ensuring 100% code coverage.
func TestPrintGreenBlueCyanPart_ShortLine(t *testing.T) {
	// Arrange
	t.Parallel()

	var writer bytes.Buffer
	shortLine := "short"

	// Act
	printGreenBlueCyanPart(&writer, shortLine)

	// Assert
	output := writer.String()
	if !strings.Contains(output, shortLine) {
		t.Errorf("Expected output to contain %q, got %q", shortLine, output)
	}
}

// TestPrintGreenBlueCyanPart_LongLine tests the normal case where the line
// is longer than cyanStartIndex.
func TestPrintGreenBlueCyanPart_LongLine(t *testing.T) {
	// Arrange
	t.Parallel()

	var writer bytes.Buffer
	longLine := "this is a very long line that exceeds the cyan start index of 38 characters"

	// Act
	printGreenBlueCyanPart(&writer, longLine)

	// Assert
	output := writer.String()
	if !strings.Contains(output, longLine) {
		t.Errorf("Expected output to contain %q, got %q", longLine, output)
	}
}

// TestPrintGreenCyanPart_ShortLine tests the edge case where the line
// is shorter than greenCyanSplitIndex, ensuring 100% code coverage.
func TestPrintGreenCyanPart_ShortLine(t *testing.T) {
	// Arrange
	t.Parallel()

	var writer bytes.Buffer
	shortLine := "short"

	// Act
	printGreenCyanPart(&writer, shortLine)

	// Assert
	output := writer.String()
	if !strings.Contains(output, shortLine) {
		t.Errorf("Expected output to contain %q, got %q", shortLine, output)
	}
}

// TestPrintGreenCyanPart_LongLine tests the normal case where the line
// is longer than greenCyanSplitIndex.
func TestPrintGreenCyanPart_LongLine(t *testing.T) {
	// Arrange
	t.Parallel()

	var writer bytes.Buffer
	longLine := "this is a long line that exceeds the green cyan split index"

	// Act
	printGreenCyanPart(&writer, longLine)

	// Assert
	output := writer.String()
	if !strings.Contains(output, longLine) {
		t.Errorf("Expected output to contain %q, got %q", longLine, output)
	}
}