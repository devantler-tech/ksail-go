package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewStatusCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewStatusCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "status" {
		t.Fatalf("expected Use to be 'status', got %q", cmd.Use)
	}

	if cmd.Short != "Show status of the Kubernetes cluster" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
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
