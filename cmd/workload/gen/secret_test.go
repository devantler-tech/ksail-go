package gen //nolint:testpackage // Tests need access to unexported helpers

import (
	"bytes"
	"fmt"
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
	cmd.SetArgs(
		[]string{
			"generic",
			"test-secret",
			"--from-literal=username=admin",
			"--from-literal=password=<PASSWORD>", // nosec G101 - test placeholder
		},
	)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen secret generic to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestGenSecretTLS tests generating a TLS secret manifest.
//
//nolint:paralleltest,gosec // Snapshot tests should not run in parallel; test uses hardcoded RSA key
func TestGenSecretTLS(t *testing.T) {
	// Create temporary cert and key files for testing
	tmpDir := t.TempDir()
	certFile := tmpDir + "/tls.crt"
	keyFile := tmpDir + "/tls.key"

	// Create valid self-signed test certificate and key
	testCert := `-----BEGIN CERTIFICATE-----
MIICEjCCAXsCAg36MA0GCSqGSIb3DQEBBQUAMIGbMQswCQYDVQQGEwJKUDEOMAwG
A1UECBMFVG9reW8xEDAOBgNVBAcTB0NodW8ta3UxETAPBgNVBAoTCEZyYW5rNERE
MRgwFgYDVQQLEw9XZWJDZXJ0IFN1cHBvcnQxGDAWBgNVBAMTD0ZyYW5rNEREIFdl
YiBDQTEjMCEGCSqGSIb3DQEJARYUc3VwcG9ydEBmcmFuazRkZC5jb20wHhcNMTIw
ODIyMDUyNjU0WhcNMTcwODIxMDUyNjU0WjBKMQswCQYDVQQGEwJKUDEOMAwGA1UE
CAwFVG9reW8xETAPBgNVBAoMCEZyYW5rNEREMRgwFgYDVQQDDA93d3cuZXhhbXBs
ZS5jb20wXDANBgkqhkiG9w0BAQEFAANLADBIAkEAm/xmkHmEQrurE/0re/jeFRLl
8ZPjBop7uLHhnia7lQG/5zDtZIUC3RVpqDSwBuw/NTweGyuP+o8AG98HxqxTBwID
AQABMA0GCSqGSIb3DQEBBQUAA4GBABS2TLuBeTPmcaTaUW/LCB2NYOy8GMdzR1mx
8iBIu2H6/E2tiY3RIevV2OW61qY2/XRQg7YPxx3ffeUugX9F4J/iPnnu1zAxzyYw
m+DF+S9yYG6bGbxw/+7L4+0D3S+J4wcYXUYZ0nJVl8C3GyPl5vk8rMOQgxNQoEb9
GcLt1B4m
-----END CERTIFICATE-----`

	// nosec G101 - test RSA private key for certificate generation
	testKey := `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAJv8ZpB5hEK7qxP9K3v43hUS5fGT4waKe7ix4Z4mu5UBv+cw7WSF
At0Vaag0sAbsPzU8HhsrL/qPABvfB8asUwcCAwEAAQJAWVu8I8S/K5Cv7Gr8TlrB
oP0FYtL0gRNWV24BYxe+oq9zL0B1FNw2IQ/r7cW1zQBjxAHdGFgV6V8sQiHUKxfD
QQIhAPfqTEXp3p6DOr6sGJRe6ggFVfJ8I6LLV2rUwqWFCJRNAiEAoLz3t3T7qhPL
g8Lh6lKlFzHE/6KKq+w8JHWGDX1wfHsCIBqKX1GzIgApxF0cqPWFR0Xw4tD7rCqU
aHxIZPwu+0ZtAiEAhHLYxzp9JLWa7LxGbdPqrL+LRCBQDJqhBRhU7vc1H5MCIDCm
gE4Q8M0VgVhGPKFjYfxHUETdC8OPi6EvRnVOB3Wv
-----END RSA PRIVATE KEY-----`

	err := writeFile(certFile, testCert)
	if err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	err = writeFile(keyFile, testKey)
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
		"--docker-password=<PASSWORD>", // nosec G101 - test placeholder
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
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
