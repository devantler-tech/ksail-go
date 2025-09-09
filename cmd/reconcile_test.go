package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
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

	if cmd.Short != "Reconcile workloads in the Kubernetes cluster" {
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

	got := out.String()

	expected := "✔ Workloads reconciled successfully (stub implementation)\n"

	if got != expected {
		t.Fatalf("expected output %q, got %q", expected, got)
	}
}
