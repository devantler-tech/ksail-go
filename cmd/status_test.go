//nolint:dupl // Test files naturally have similar patterns for different commands
package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewStatusCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "status",
		ExpectedUse:    "status",
		ExpectedShort:  "Show status of the Kubernetes cluster",
		NewCommandFunc: cmd.NewStatusCmd,
	})
}

func TestStatusCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStatusCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestStatusCmd_Help(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStatusCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
