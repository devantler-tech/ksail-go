//nolint:dupl // Test files naturally have similar patterns for different commands
package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewReconcileCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "reconcile",
		ExpectedUse:    "reconcile",
		ExpectedShort:  "Reconcile workloads in the cluster",
		NewCommandFunc: cmd.NewReconcileCmd,
	})
}

func TestReconcileCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewReconcileCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestReconcileCmd_Help(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewReconcileCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
