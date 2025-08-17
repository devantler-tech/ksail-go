package asciiart_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/asciiart"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPrintKSailLogo(test *testing.T) {
	// Arrange
	test.Parallel()

	var writer bytes.Buffer

	// Act
	asciiart.PrintKSailLogo(&writer)

	// Assert
	snaps.MatchSnapshot(test, writer.String())
}
