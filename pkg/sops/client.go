// Package sops provides a sops client implementation for executing sops commands.
//
// # Implementation Note
//
// This package uses exec to run the sops binary rather than importing sops as a Go library.
// This is intentional because:
//
//  1. SOPS uses urfave/cli v1, which is incompatible with Cobra
//  2. SOPS doesn't export an app builder - all commands are defined in a 2500+ line main.go
//  3. The only stable API SOPS provides is for decryption (github.com/getsops/sops/v3/decrypt)
//  4. Wrapping urfave/cli apps in Cobra requires significant code duplication and is fragile
//
// A generic urfave/cli to Cobra wrapper is provided in pkg/cliwrapper for other use cases,
// but for SOPS specifically, the exec approach is cleaner and more maintainable.
package sops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// ErrSopsExecution is returned when sops command execution fails.
var ErrSopsExecution = errors.New("sops command execution failed")

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
		RunE: func(_ *cobra.Command, args []string) error {
			return c.executeSops(args)
		},
		SilenceUsage: true,
	}

	return cmd
}

// executeSops runs the sops binary with the provided arguments.
func (c *Client) executeSops(args []string) error {
	ctx := context.Background()
	sopsCmd := exec.CommandContext(ctx, "sops", args...)
	sopsCmd.Stdin = os.Stdin
	sopsCmd.Stdout = os.Stdout
	sopsCmd.Stderr = os.Stderr

	err := sopsCmd.Run()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSopsExecution, err)
	}

	return nil
}
