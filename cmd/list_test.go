package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNewListCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewListCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "list" {
		t.Fatalf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short != "List Kubernetes clusters" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestListCmd_Execute_Default(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestListCmd_Execute_All(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestListCmd_Help(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewListCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestListCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewListCmd()

	// Act & Assert
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("expected all flag to exist")
	}

	if allFlag.DefValue != "false" {
		t.Fatalf("expected all flag default to be 'false', got %q", allFlag.DefValue)
	}
}
