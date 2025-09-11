//nolint:dupl // Test files naturally have similar patterns for different CLI commands
package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewReconcileCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewReconcileCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "reconcile" {
		t.Fatalf("expected Use to be 'reconcile', got %q", cmd.Use)
	}

	if cmd.Short != "Reconcile workloads in the cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
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
