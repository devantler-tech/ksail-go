// Package sops provides a sops client implementation that delegates to the sops binary.
//
// # Implementation Note
//
// This package wraps the sops binary execution, passing through all commands and flags.
// This approach ensures complete SOPS feature parity while maintaining clean integration
// with KSail's Cobra-based CLI structure.
//
// # Dependencies
//
// This command requires the sops binary to be installed and available in the system PATH.
// Install sops from: https://github.com/getsops/sops
package sops

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// Client wraps sops command functionality.
type Client struct{}

// NewClient creates a new sops client instance.
func NewClient() *Client {
	return &Client{}
}

// CreateCipherCommand creates a cipher command that delegates to the sops binary.
func (c *Client) CreateCipherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cipher",
		Short: "Manage encrypted files",
		Long: `Cipher command provides encryption and decryption functionality
for managing encrypted files.

ksail cipher supports multiple key management systems:
  - age recipients (-a, --age)
  - PGP fingerprints (-p, --pgp)
  - AWS KMS (-k, --kms)
  - GCP KMS (--gcp-kms)
  - Azure Key Vault (--azure-kv)
  - HashiCorp Vault (--hc-vault-transit)

Common operations:
  ksail cipher --encrypt file.yaml         # Encrypt a file
  ksail cipher --decrypt file.yaml         # Decrypt a file
  ksail cipher --rotate file.yaml          # Rotate data encryption key
  ksail cipher --set '["key"] value' file  # Set a value
  ksail cipher --edit file.yaml            # Edit encrypted file

Dependencies:
  This command requires the 'sops' binary to be installed.
  Install from: https://github.com/getsops/sops`,
		RunE:                       c.handleCipherRunE,
		SilenceUsage:               true,
		DisableFlagParsing:         true, // Pass all flags directly to sops
		DisableFlagsInUseLine:      true,
		SuggestionsMinimumDistance: 2, //nolint:mnd // Standard cobra suggestion distance
	}

	return cmd
}

// handleCipherRunE executes the sops binary with all provided arguments.
func (c *Client) handleCipherRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Check if sops is available
	sopsPath, err := exec.LookPath("sops")
	if err != nil {
		return fmt.Errorf(
			"sops binary not found in PATH: %w\n\nPlease install sops from: https://github.com/getsops/sops",
			err,
		)
	}

	// Create command to execute sops with all provided arguments
	//nolint:gosec // This is intentional - we're delegating to the sops binary with user-provided args
	sopsCmd := exec.CommandContext(ctx, sopsPath, args...)
	sopsCmd.Stdin = os.Stdin
	sopsCmd.Stdout = cmd.OutOrStdout()
	sopsCmd.Stderr = cmd.ErrOrStderr()

	// Execute sops
	err = sopsCmd.Run()
	if err != nil {
		return fmt.Errorf("sops execution failed: %w", err)
	}

	return nil
}
