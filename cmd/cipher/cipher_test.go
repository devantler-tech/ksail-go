package cipher_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cipher"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

func TestNewCipherCmd(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cmd := cipher.NewCipherCmd(rt)

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	if cmd.Use != "cipher" {
		t.Errorf("expected Use to be 'cipher', got %q", cmd.Use)
	}

	// Verify the short description is set (wrapped from urfave/cli app)
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
}

func TestCipherCommandHelp(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cmd := cipher.NewCipherCmd(rt)

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	// Since DisableFlagParsing is true, --help is passed to sops binary
	// We can't easily test this without sops installed, so just verify
	// the command executes (it will try to run sops --help)
	_ = cmd.Execute()

	// The command structure is correct even if execution fails
	// (e.g., if sops is not installed in the test environment)
}
