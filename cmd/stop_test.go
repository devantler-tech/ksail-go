//nolint:dupl // Test files naturally have similar patterns for different commands
package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewStopCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "stop",
		ExpectedUse:    "stop",
		ExpectedShort:  "Stop the Kubernetes cluster",
		NewCommandFunc: cmd.NewStopCmd,
	})
}

func TestStopCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStopCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestStopCmd_Help(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStopCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
