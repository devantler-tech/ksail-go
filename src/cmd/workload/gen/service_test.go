package gen //nolint:testpackage // Tests need access to unexported helpers

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenServiceClusterIP tests generating a ClusterIP service manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenServiceClusterIP(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewServiceCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"clusterip", "test-svc", "--tcp=80:8080"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen service clusterip to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenServiceNodePort tests generating a NodePort service manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenServiceNodePort(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewServiceCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"nodeport", "test-svc", "--tcp=80:8080"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen service nodeport to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenServiceLoadBalancer tests generating a LoadBalancer service manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenServiceLoadBalancer(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewServiceCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"loadbalancer", "test-svc", "--tcp=80:8080"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen service loadbalancer to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenServiceExternalName tests generating an ExternalName service manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenServiceExternalName(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewServiceCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"externalname", "test-svc", "--external-name=example.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen service externalname to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
