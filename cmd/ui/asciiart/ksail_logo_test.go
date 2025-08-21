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
