// Package cipher provides the cipher command for AES encryption/decryption.
package cipher

import (
	"fmt"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewCipherCmd creates the cipher command with encrypt/decrypt subcommands.
func NewCipherCmd(_ *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cipher",
		Short: "Encrypt and decrypt files using AES-256-GCM",
		Long: `Cipher command provides AES-256-GCM encryption and decryption functionality
using SOPS cryptographic primitives.

Subcommands:
  encrypt  Encrypt a file with AES-256-GCM
  decrypt  Decrypt an AES-256-GCM encrypted file

The encrypted format is compatible with SOPS:
  ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]`,
		RunE:         handleCipherRunE,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewEncryptCmd())
	cmd.AddCommand(NewDecryptCmd())

	return cmd
}

//nolint:gochecknoglobals // Injected for testability to simulate help failures.
var helpRunner = func(cmd *cobra.Command) error {
	return cmd.Help()
}

func handleCipherRunE(cmd *cobra.Command, _ []string) error {
	// Display help when cipher is called without subcommands
	err := helpRunner(cmd)
	if err != nil {
		return fmt.Errorf("displaying cipher command help: %w", err)
	}

	return nil
}
