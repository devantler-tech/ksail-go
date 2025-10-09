// Package builder provides a builder for creating SOPS urfave/cli applications.
package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/getsops/sops/v3/version" //nolint:depguard // Required for sops version info
	"github.com/urfave/cli"              //nolint:depguard // This package wraps urfave/cli apps
)

// NewSopsApp creates a urfave/cli app that wraps SOPS functionality.
// This app can be wrapped with pkg/cliwrapper to integrate with Cobra.
//
// Note: This is a pragmatic implementation that delegates to the sops binary
// for actual operations, wrapped in a urfave/cli structure for compatibility.
func NewSopsApp() *cli.App {
	app := cli.NewApp()
	app.Name = "cipher"
	app.Usage = "sops - encrypted file editor with AWS KMS, GCP KMS, Azure Key Vault, age, and GPG support"
	app.Version = version.Version
	app.Authors = []cli.Author{
		{Name: "CNCF Maintainers"},
	}

	app.UsageText = `Cipher command provides access to all SOPS (Secrets OPerationS) functionality
for encrypting and decrypting files.

To encrypt or decrypt a document with AWS KMS, specify the KMS ARN
in the -k flag or in the SOPS_KMS_ARN environment variable.

To encrypt or decrypt a document with GCP KMS, specify the
GCP KMS resource ID in the --gcp-kms flag or in the SOPS_GCP_KMS_IDS
environment variable.

To encrypt or decrypt using age, specify the recipient in the -a flag,
or in the SOPS_AGE_RECIPIENTS environment variable.

To encrypt or decrypt using PGP, specify the PGP fingerprint in the
-p flag or in the SOPS_PGP_FP environment variable.`

	// Default action - delegate to sops binary
	app.Action = func(c *cli.Context) error {
		return executeSopsBinary(c.Args())
	}

	// Define subcommands that delegate to sops
	app.Commands = createSopsCommands()

	return app
}

//nolint:funlen // Command definitions require length for clarity
func createSopsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "encrypt",
			Usage: "encrypt a file, and output the results to stdout",
			Action: func(c *cli.Context) error {
				args := append([]string{"encrypt"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "decrypt",
			Usage: "decrypt a file, and output the results to stdout",
			Action: func(c *cli.Context) error {
				args := append([]string{"decrypt"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "rotate",
			Usage: "generate a new data encryption key and reencrypt all values with the new key",
			Action: func(c *cli.Context) error {
				args := append([]string{"rotate"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "edit",
			Usage: "edit an encrypted file",
			Action: func(c *cli.Context) error {
				args := append([]string{"edit"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "set",
			Usage: "set a specific key or branch in the input document",
			Action: func(c *cli.Context) error {
				args := append([]string{"set"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "unset",
			Usage: "unset a specific key or branch in the input document",
			Action: func(c *cli.Context) error {
				args := append([]string{"unset"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "updatekeys",
			Usage: "update the keys of SOPS files using the config file",
			Action: func(c *cli.Context) error {
				args := append([]string{"updatekeys"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
		{
			Name:  "groups",
			Usage: "modify the groups on a SOPS file",
			Action: func(c *cli.Context) error {
				args := append([]string{"groups"}, c.Args()...)

				return executeSopsBinary(args)
			},
			SkipFlagParsing: true,
		},
	}
}

// executeSopsBinary executes the sops binary with the provided arguments.
func executeSopsBinary(args []string) error {
	ctx := context.Background()
	sopsCmd := exec.CommandContext(ctx, "sops", args...)
	sopsCmd.Stdin = os.Stdin
	sopsCmd.Stdout = os.Stdout
	sopsCmd.Stderr = os.Stderr

	err := sopsCmd.Run()
	if err != nil {
		return fmt.Errorf("sops command execution failed: %w", err)
	}

	return nil
}
