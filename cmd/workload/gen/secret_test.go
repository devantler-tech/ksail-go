package gen //nolint:testpackage // Tests need access to unexported helpers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

const (
	testTLSCert = `-----BEGIN CERTIFICATE-----
MIIDDTCCAfWgAwIBAgIUQg2thFOmdEGn073/v2LH13oF0bIwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjUxMTAzMTk0MjIxWhcNMzUx
MTAxMTk0MjIxWjAWMRQwEgYDVQQDDAtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAKmOXZrM0wpyfNPu40G2gsXXNPknjzB/RjtkDY8N
Ni/YfBK8MR41+ttjF3GRt2nkIDZ7gxMfDYBbQ5z4RQgarPLc+ADw+TnQ8aWAJGHN
TrUlqUsQaYMPi1nU066P0ctFvS0ezzJ9QblnJLDbhobvykK5wXp9pWGFvCyGBSGA
LoH3S1dZpRNazl7YexHVzo4aqDzu9B1mBDm9FP1aPgfCYX+o9ZfHFpFJkGT8Uutn
EYXSb/zedRHzYw2ya23zqCZ8fGlxWYD4+jwJyogJ2P5hPwZQ2t0biDNWhi0B5VuL
CCmNKEZsRQkOhWHH6rfmm1XqgM8wRIii+o3B4I3/9jbBGF0CAwEAAaNTMFEwHQYD
VR0OBBYEFL8A6pmICsjO1G+Y+UQ2ySAiCxK7MB8GA1UdIwQYMBaAFL8A6pmICsjO
1G+Y+UQ2ySAiCxK7MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AErf6LUuXvuL/GLixJjOBADpM9Leju7dbB2t+sDh+kIPDpsHMj4EjPshSisBm26m
emE6geKA0vjD4fI8RL/kIvlPzPwojBDkbqOzNIxsAUF+7jlOxabuCmsQqpjZf4I7
zxomDNeSDndqUgcJIf/HjxyWK5Fi4N1wQuoid375EEixavXmzBIQvvXD2qT44rGY
vneGmModP5G4mcUIdNAd1oQoGYYKFpDPu7a1DiBGWTWb8sifjBwWjHhC1IsHMKg1
L1SjRMzmGtmQ7ckyJjq/cDDcvqei6tPKhN6oLjVezyfgb/j3feQM34RxOOMlm7IV
ZQL4GfN3z39LdLpniz7OuqU=
-----END CERTIFICATE-----
`

	// nosec G101 - Private key used solely for test fixtures.
	testTLSKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCpjl2azNMKcnzT
7uNBtoLF1zT5J48wf0Y7ZA2PDTYv2HwSvDEeNfrbYxdxkbdp5CA2e4MTHw2AW0Oc
+EUIGqzy3PgA8Pk50PGlgCRhzU61JalLEGmDD4tZ1NOuj9HLRb0tHs8yfUG5ZySw
24aG78pCucF6faVhhbwshgUhgC6B90tXWaUTWs5e2HsR1c6OGqg87vQdZgQ5vRT9
Wj4HwmF/qPWXxxaRSZBk/FLrZxGF0m/83nUR82MNsmtt86gmfHxpcVmA+Po8CcqI
Cdj+YT8GUNrdG4gzVoYtAeVbiwgpjShGbEUJDoVhx+q35ptV6oDPMESIovqNweCN
//Y2wRhdAgMBAAECggEAVINLkM8rGff61EAsMiLgh/A+zTm0m3206gFy6KyzJ6IG
Jeh7qw1I3nVDyC3TeApnLADgUnWV6zaSOvlcny98qQkO7JkwAGtvJwj6GW2WH6CI
A4xIqzTiRoJYiJfTADjglE7ZA9d/HQSWOzkQks2OyTeBgqaB+lwIcUDT6eDUTZ6/
JtVY5EZmn+JEKylHJznnoEIIEyjYjJED33bQX4GszKojrD2tNY3ASgKhi+6cienm
PHt0I8l0EdoNUv1tCzxzZuyhqush9S4HY+EGcyFLj5drYzVDG8L8+EOndJd6cZ3L
IJGQDgUigKGKPwAR4+XnvKJMIBNBIZDzpWBprxWKmQKBgQDh0S+Nl/hsQjQsTyXt
qnRFYBTXneQrmCm/YHSBl4UGXK4z+XxqxGJEe8+0OVok7TIkSQyIF0tvcXJWjwUd
H9VYx1CLybldCdXlzSf4uM3inlWsgBRK/Ft0IDyw13bZ0L5XuY96O82ZWDR6/7gA
J2Sf1nVUqibBt9JRdXBWdO2HTwKBgQDAOBbguICAUCOQD959Z1063Kai0ChF1cXM
8wkA3iDTynJUOlxW+tn9n0PiHqcmjcLC3gQrxjb53qC7iakXnVw+rO1Xhfa9XzfK
slI5JBUueZtD+P45ZeRCaM1LhevIFSBoFOPPJYzninYOjawZZyttn7vzKFxVAr8a
DqOZjeO6kwKBgQDQVXzoxi8kOcQGqRLV/O9+XdF8x6eNbLn/XQ6/zLmmj/UL8H2P
xxTeF9gdbtgyvz8GaPqNx+gJrgGNyC8wmoDrgh9WiEpigsN7WtYoyt7v16I1Hoka
UU5SibdUc8Sr2cDyEDlFzUy2z8DDRY9NXQqhyGrBLKXLDTuVeaKlsQS/UwKBgQCd
KH7UByXRQzSAYekoIO3h5Ww86/IxfuH1erPeyL6QSxKE+R5sYzb+HUx0QVmqtPcL
OliwraRfUX2bN6dPznIQMHTxPW+KT6KfEIMXgv/qerTOs3Kv3TXuch9/4yPu+A8B
6iqEQBBfcx6pMX4HWwnv3EzgNxyeyNsUY+mw74jFDwKBgHT05YAt2RslNS4fmDvb
WhPIJGokCb27z32bH5jVVfr2Uq/GfWiIRY01KTWFLKSZseQA8SlOZ52q7NAckan5
ptRP8mvEaJVFBiIf95JlkTp76qUDLrEhI2ALJx1JjVx4H1M3Jjoeelm1qKGesaYz
H25Qf9zEQeJJCcSQPZ+iipaX
-----END PRIVATE KEY-----
`
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
			"--from-literal=key1=value1",
			"--from-literal=key2=value2",
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
	certFile, keyFile := createTLSFixtures(t)

	rt := newTestRuntime()
	cmd := NewSecretCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"tls", "test-tls-secret", "--cert=" + certFile, "--key=" + keyFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen secret tls to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

// TestTLSFixturesParse verifies the TLS fixtures are valid.
func TestTLSFixturesParse(t *testing.T) {
	// Validate the in-memory PEM pair first to ensure test data stays consistent.
	if _, err := tls.X509KeyPair([]byte(testTLSCert), []byte(testTLSKey)); err != nil {
		t.Fatalf("TLS fixtures should parse in-memory: %v", err)
	}

	certPath, keyPath := createTLSFixtures(t)
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		t.Fatalf("TLS fixtures should parse from disk: %v", err)
	}
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
		"--docker-username=testuser",
		"--docker-password=testpass123",
		"--docker-email=testuser@example.com",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen secret docker-registry to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}

func createTLSFixtures(t *testing.T) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")

	err := writeFile(certFile, testTLSCert)
	if err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	err := writeFile(keyFile, testTLSKey)
	if err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	return certFile, keyFile
}

// writeFile is a helper to write test files.
func writeFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
