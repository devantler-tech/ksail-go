package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewStartCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewStartCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "start" {
		t.Fatalf("expected Use to be 'start', got %q", cmd.Use)
	}

	if cmd.Short != "Start a stopped cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestStartCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStartCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestStartCmd_Help(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewStartCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
