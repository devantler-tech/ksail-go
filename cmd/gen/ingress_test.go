package gen //nolint:testpackage // Tests need access to unexported helpers

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenIngressSimple tests generating a simple ingress manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenIngressSimple(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewIngressCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-ingress", "--rule=example.com/*=svc:80"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen ingress to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenIngressWithTLS tests generating an ingress manifest with TLS.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenIngressWithTLS(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewIngressCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-ingress-tls", "--rule=secure.example.com/*=svc:443,tls=my-tls-secret"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen ingress with TLS to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenIngressMultipleRules tests generating an ingress with multiple rules.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenIngressMultipleRules(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewIngressCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"test-ingress-multi",
		"--rule=api.example.com/*=api-svc:8080",
		"--rule=web.example.com/*=web-svc:80",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen ingress with multiple rules to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
