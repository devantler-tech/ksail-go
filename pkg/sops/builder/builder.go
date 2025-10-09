// Package builder provides a builder for creating SOPS urfave/cli applications.
package builder

import (
	"errors"
	"fmt"
	"os"

	"github.com/getsops/sops/v3"         //nolint:depguard // Required for SOPS operations
	"github.com/getsops/sops/v3/age"     //nolint:depguard // Required for age encryption
	"github.com/getsops/sops/v3/pgp"     //nolint:depguard // Required for PGP encryption
	"github.com/getsops/sops/v3/version" //nolint:depguard // Required for sops version info
	"github.com/urfave/cli"              //nolint:depguard // This package wraps urfave/cli apps

	"github.com/devantler-tech/ksail-go/pkg/sops/operations"
)

var (
	// ErrNotImplemented is returned when a command is not yet implemented.
	ErrNotImplemented = errors.New("command not yet implemented with Go libraries")
	// ErrNoInputFile is returned when no input file is specified.
	ErrNoInputFile = errors.New("no input file specified")
	// ErrNoEncryptionKeys is returned when no encryption keys are specified.
	ErrNoEncryptionKeys = errors.New("no encryption keys specified (use --age or --pgp)")
)

// NewSopsApp creates a urfave/cli app that wraps SOPS functionality.
// This app can be wrapped with pkg/cliwrapper to integrate with Cobra.
//
// Note: This implementation uses SOPS Go libraries directly for operations.
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

	// Default action
	app.Action = cli.ShowAppHelp

	// Define subcommands that use SOPS libraries
	app.Commands = createSopsCommands()

	return app
}

//nolint:funlen // Command definitions require length for clarity
func createSopsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "encrypt",
			Usage: "encrypt a file, and output the results to stdout",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "age, a",
					Usage: "comma separated list of age recipients",
				},
				cli.StringFlag{
					Name:  "pgp, p",
					Usage: "comma separated list of PGP fingerprints",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
			},
			Action: handleEncrypt,
		},
		{
			Name:  "decrypt",
			Usage: "decrypt a file, and output the results to stdout",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
			},
			Action: handleDecrypt,
		},
		{
			Name:  "rotate",
			Usage: "generate a new data encryption key and reencrypt all values with the new key",
			Action: func(_ *cli.Context) error {
				return ErrNotImplemented
			},
		},
		{
			Name:  "edit",
			Usage: "edit an encrypted file",
			Action: func(_ *cli.Context) error {
				return ErrNotImplemented
			},
		},
		{
			Name:  "set",
			Usage: "set a specific key or branch in the input document",
			Action: func(_ *cli.Context) error {
				return ErrNotImplemented
			},
		},
		{
			Name:  "unset",
			Usage: "unset a specific key or branch in the input document",
			Action: func(_ *cli.Context) error {
				return ErrNotImplemented
			},
		},
		{
			Name:  "updatekeys",
			Usage: "update the keys of SOPS files using the config file",
			Action: func(_ *cli.Context) error {
				return ErrNotImplemented
			},
		},
		{
			Name:  "groups",
			Usage: "modify the groups on a SOPS file",
			Action: func(_ *cli.Context) error {
				return ErrNotImplemented
			},
		},
	}
}

func handleEncrypt(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")

	// Parse key groups from flags
	keyGroups, err := parseKeyGroups(cliCtx)
	if err != nil {
		return fmt.Errorf("failed to parse key groups: %w", err)
	}

	if len(keyGroups) == 0 {
		return ErrNoEncryptionKeys
	}

	// Encrypt file
	err = operations.EncryptFileToWriter(inputFile, keyGroups, outputFormat, os.Stdout)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return nil
}

func handleDecrypt(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")

	// Decrypt file
	err := operations.DecryptFileToWriter(inputFile, outputFormat, os.Stdout)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return nil
}

func parseKeyGroups(cliCtx *cli.Context) ([]sops.KeyGroup, error) {
	var keyGroup sops.KeyGroup

	// Parse age recipients
	ageRecipients := cliCtx.String("age")
	if ageRecipients != "" {
		masterKey := &age.MasterKey{
			Recipient: ageRecipients,
		}
		keyGroup = append(keyGroup, masterKey)
	}

	// Parse PGP fingerprints
	pgpFingerprints := cliCtx.String("pgp")
	if pgpFingerprints != "" {
		masterKey := pgp.NewMasterKeyFromFingerprint(pgpFingerprints)
		keyGroup = append(keyGroup, masterKey)
	}

	if len(keyGroup) == 0 {
		return nil, nil
	}

	return []sops.KeyGroup{keyGroup}, nil
}
