// Package asciiart provides ASCII art printing functionality for KSail.
package asciiart

import (
	_ "embed"
	"io"
	"strings"

	"github.com/fatih/color"
)

//go:embed ksail_logo.txt
var ksailLogo string

const (
	greenCyanSplitIndex = 32
	blueStartIndex      = 37
	cyanStartIndex      = 38
)

// PrintKSailLogo displays the KSail ASCII art with colored formatting.
func PrintKSailLogo(writer io.Writer) {
	const yellowLines = 4

	lines := strings.Split(ksailLogo, "\n")

	for index, line := range lines {
		switch {
		case index < yellowLines:
			printlnc(writer, colorYellow(), line)
		case index == yellowLines:
			printlnc(writer, colorBlue(), line)
		case index > yellowLines && index < 7:
			printGreenBlueCyanPart(writer, line)
		case index > 6 && index < len(lines)-2:
			printGreenCyanPart(writer, line)
		default:
			printlnc(writer, colorBlue(), line)
		}
	}
}

// --- internals ---

// small helpers to reduce repetition.
func printc(out io.Writer, c *color.Color, s string)   { _, _ = c.Fprint(out, s) }
func printlnc(out io.Writer, c *color.Color, s string) { _, _ = c.Fprintln(out, s) }

// color constructors (avoid package globals for linter compliance).
func colorYellow() *color.Color { return color.New(color.Bold, color.FgYellow) }
func colorBlue() *color.Color   { return color.New(color.Bold, color.FgBlue) }
func colorGreen() *color.Color  { return color.New(color.Bold, color.FgGreen) }
func colorCyan() *color.Color   { return color.New(color.FgCyan) }

func printGreenBlueCyanPart(out io.Writer, line string) {
	if len(line) >= cyanStartIndex {
		printc(out, colorGreen(), line[:greenCyanSplitIndex])
		printc(out, colorCyan(), line[greenCyanSplitIndex:blueStartIndex])
		printc(out, colorBlue(), line[blueStartIndex:cyanStartIndex])
		printlnc(out, colorCyan(), line[cyanStartIndex:])
	} else {
		printlnc(out, colorGreen(), line)
	}
}

func printGreenCyanPart(out io.Writer, line string) {
	if len(line) >= greenCyanSplitIndex {
		printc(out, colorGreen(), line[:greenCyanSplitIndex])
		printlnc(out, colorCyan(), line[greenCyanSplitIndex:])
	} else {
		printlnc(out, colorGreen(), line)
	}
}
