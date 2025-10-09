// Package sops provides a sops client implementation for executing sops commands.
package sops

import (
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

// CreateCipherCommand creates a cipher command that delegates all subcommands and flags to sops.
func (c *Client) CreateCipherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cipher",
		Short: "Manage encryption and decryption with SOPS",
		Long: "Cipher command provides access to all SOPS (Secrets OPerationS) functionality " +
			"for encrypting and decrypting files. All subcommands and flags are passed directly to sops.",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.executeSops(args)
		},
		SilenceUsage: true,
	}

	return cmd
}

// executeSops runs the sops binary with the provided arguments.
func (c *Client) executeSops(args []string) error {
	sopsCmd := exec.Command("sops", args...)
	sopsCmd.Stdin = os.Stdin
	sopsCmd.Stdout = os.Stdout
	sopsCmd.Stderr = os.Stderr

	return sopsCmd.Run()
}
