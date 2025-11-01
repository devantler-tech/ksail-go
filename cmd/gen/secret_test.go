package gen //nolint:testpackage // Tests need access to unexported helpers

import (
	"bytes"
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenSecretGeneric tests generating a generic secret manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenSecretGeneric(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewSecretCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"generic", "test-secret", "--from-literal", "username=admin", "--from-literal", "password=secret123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen secret generic to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenSecretTLS tests generating a TLS secret manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenSecretTLS(t *testing.T) {
	// Create temporary cert and key files for testing
	tmpDir := t.TempDir()
	certFile := tmpDir + "/tls.crt"
	keyFile := tmpDir + "/tls.key"

	// Write dummy cert and key
	err := writeFile(certFile, "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n")
	if err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	err = writeFile(keyFile, "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n")
	if err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	rt := newTestRuntime()
	cmd := NewSecretCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"tls", "test-tls-secret", "--cert=" + certFile, "--key=" + keyFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen secret tls to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenSecretDockerRegistry tests generating a docker-registry secret manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenSecretDockerRegistry(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewSecretCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"docker-registry",
		"test-docker-secret",
		"--docker-server=https://registry.example.com",
		"--docker-username=user",
		"--docker-password=pass",
		"--docker-email=user@example.com",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen secret docker-registry to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// writeFile is a helper to write test files.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
