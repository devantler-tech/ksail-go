// Package asciiart provides ASCII art printing functionality for KSail.
package asciiart

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed ksail.txt
var ksailLogo string

// PrintKSailLogo displays the KSail ASCII art with colored formatting.
func PrintKSailLogo() {
	const yellowLines = 4

	lines := strings.Split(ksailLogo, "\n")

	for index, line := range lines {
		switch {
		case index < yellowLines:
			printYellowPart(line)
		case index == yellowLines:
			printBluePart(line)
		case index > yellowLines && index < 7:
			printGreenBlueCyanPart(line)
		case index > 6 && index < len(lines)-2:
			printGreenCyanPart(line)
		default:
			printBluePart(line)
		}
	}
}

func printYellowPart(line string) {
	fmt.Println("\x1b[1;33m" + line + "\x1b[0m")
}

func printBluePart(line string) {
	fmt.Println("\x1b[1;34m" + line + "\x1b[0m")
}

func printGreenBlueCyanPart(line string) {
	charThirtyEight := 38
	if len(line) >= charThirtyEight {
		fmt.Print("\x1b[1;32m" + line[:32] + "\x1b[0m")
		fmt.Print("\x1B[36m" + line[32:37] + "\x1b[0m")
		fmt.Print("\x1b[1;34m" + line[37:charThirtyEight] + "\x1b[0m")
		fmt.Println("\x1B[36m" + line[38:] + "\x1b[0m")
	} else {
		fmt.Println("\x1b[1;32m" + line + "\x1b[0m")
	}
}

const greenCyanSplitIndex = 32

func printGreenCyanPart(line string) {
	if len(line) >= greenCyanSplitIndex {
		fmt.Print("\x1b[1;32m" + line[:greenCyanSplitIndex] + "\x1b[0m")
		fmt.Println("\x1B[36m" + line[greenCyanSplitIndex:] + "\x1b[0m")
	} else {
		fmt.Println("\x1b[1;32m" + line + "\x1b[0m")
	}
}
