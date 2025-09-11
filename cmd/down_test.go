//nolint:dupl // Test files naturally have similar patterns for different CLI commands
package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewDownCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewDownCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "down" {
		t.Fatalf("expected Use to be 'down', got %q", cmd.Use)
	}

	if cmd.Short != "Destroy a cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
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
