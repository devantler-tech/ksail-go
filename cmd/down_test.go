//nolint:dupl // Test files naturally have similar patterns for different commands
package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewDownCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "down",
		ExpectedUse:    "down",
		ExpectedShort:  "Destroy a cluster",
		NewCommandFunc: cmd.NewDownCmd,
	})
}

func TestDownCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewDownCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestDownCmd_Help(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewDownCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
