// Package sops provides a sops client implementation using urfave/cli app wrapping.
//
// # Implementation Note
//
// This package uses pkg/cliwrapper to wrap a urfave/cli app (created by pkg/sops/builder)
// within Cobra commands. The urfave/cli app delegates to the sops binary for actual operations
// while providing a structured command interface.
//
// This approach provides:
//  1. Integration with Cobra through pkg/cliwrapper
//  2. Structured command definitions for better help and discoverability
//  3. Delegation to sops binary for actual encryption/decryption operations
//  4. Compatibility with the full sops feature set
package sops

import (
	"github.com/devantler-tech/ksail-go/pkg/cliwrapper"
	"github.com/devantler-tech/ksail-go/pkg/sops/builder"
	"github.com/spf13/cobra"
)

// Client wraps sops command functionality.
type Client struct{}

// NewClient creates a new sops client instance.
func NewClient() *Client {
	return &Client{}
}

// CreateCipherCommand creates a cipher command that integrates SOPS via urfave/cli wrapper.
func (c *Client) CreateCipherCommand() *cobra.Command {
	// Create the SOPS urfave/cli app
	sopsApp := builder.NewSopsApp()

	// Wrap it in a Cobra command using the cliwrapper
	return cliwrapper.WrapCliApp(sopsApp)
}
