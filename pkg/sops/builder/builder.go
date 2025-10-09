// Package builder provides a builder for creating SOPS urfave/cli applications.
package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devantler-tech/ksail-go/pkg/sops/operations"
	"github.com/getsops/sops/v3"         //nolint:depguard // Required for SOPS operations
	"github.com/getsops/sops/v3/age"     //nolint:depguard // Required for age encryption
	"github.com/getsops/sops/v3/azkv"    //nolint:depguard // Required for Azure Key Vault
	"github.com/getsops/sops/v3/gcpkms"  //nolint:depguard // Required for GCP KMS
	"github.com/getsops/sops/v3/hcvault" //nolint:depguard // Required for HashiCorp Vault
	"github.com/getsops/sops/v3/kms"     //nolint:depguard // Required for AWS KMS
	"github.com/getsops/sops/v3/pgp"     //nolint:depguard // Required for PGP encryption
	"github.com/getsops/sops/v3/version" //nolint:depguard // Required for sops version info
	"github.com/urfave/cli"              //nolint:depguard // This package wraps urfave/cli apps
)

var (
	// ErrNotImplemented is returned when a command is not yet implemented.
	ErrNotImplemented = errors.New("command not yet implemented with Go libraries")
	// ErrNoInputFile is returned when no input file is specified.
	ErrNoInputFile = errors.New("no input file specified")
	// ErrNoEncryptionKeys is returned when no encryption keys are specified.
	ErrNoEncryptionKeys = errors.New("no encryption keys specified (use --age or --pgp)")
	// ErrInvalidSetArgs is returned when set command has invalid arguments.
	ErrInvalidSetArgs = errors.New("usage: set <file> <key> <value>")
	// ErrInvalidUnsetArgs is returned when unset command has invalid arguments.
	ErrInvalidUnsetArgs = errors.New("usage: unset <file> <key>")
	// ErrNoEditor is returned when EDITOR environment variable is not set.
	ErrNoEditor = errors.New("EDITOR environment variable not set")
	// ErrInvalidGroupIndex is returned when group index is invalid.
	ErrInvalidGroupIndex = errors.New("invalid group index")
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

//nolint:funlen,gocognit,maintidx,dupl,cyclop,lll // Command definitions require length and complexity for SOPS feature parity
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
					Name:  "kms, k",
					Usage: "comma separated list of KMS ARNs",
				},
				cli.StringFlag{
					Name:  "gcp-kms",
					Usage: "comma separated list of GCP KMS resource IDs",
				},
				cli.StringFlag{
					Name:  "azure-kv",
					Usage: "comma separated list of Azure Key Vault URLs",
				},
				cli.StringFlag{
					Name:  "hc-vault-transit",
					Usage: "comma separated list of HashiCorp Vault Transit URIs",
				},
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
				cli.StringFlag{
					Name:  "unencrypted-suffix",
					Usage: "override the unencrypted key suffix",
				},
				cli.StringFlag{
					Name:  "encrypted-suffix",
					Usage: "override the encrypted key suffix",
				},
				cli.StringFlag{
					Name:  "unencrypted-regex",
					Usage: "set the unencrypted key regex",
				},
				cli.StringFlag{
					Name:  "encrypted-regex",
					Usage: "set the encrypted key regex",
				},
				cli.IntFlag{
					Name:  "shamir-secret-sharing-threshold",
					Usage: "number of master keys required to retrieve the data key",
					Value: 0,
				},
			},
			Action: handleEncrypt,
		},
		{
			Name:  "decrypt",
			Usage: "decrypt a file, and output the results to stdout",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
				cli.StringFlag{
					Name:  "extract",
					Usage: "extract a specific key or branch from the input document (e.g., '[\"somekey\"][0]')",
				},
				cli.BoolFlag{
					Name:  "ignore-mac",
					Usage: "ignore Message Authentication Code during decryption",
				},
			},
			Action: handleDecrypt,
		},
		{
			Name:  "rotate",
			Usage: "generate a new data encryption key and reencrypt all values with the new key",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
				cli.StringFlag{
					Name:  "add-age",
					Usage: "add age recipient during rotation",
				},
				cli.StringFlag{
					Name:  "rm-age",
					Usage: "remove age recipient during rotation",
				},
				cli.StringFlag{
					Name:  "add-pgp",
					Usage: "add PGP fingerprint during rotation",
				},
				cli.StringFlag{
					Name:  "rm-pgp",
					Usage: "remove PGP fingerprint during rotation",
				},
				cli.StringFlag{
					Name:  "add-kms",
					Usage: "add KMS ARN during rotation",
				},
				cli.StringFlag{
					Name:  "rm-kms",
					Usage: "remove KMS ARN during rotation",
				},
				cli.StringFlag{
					Name:  "add-gcp-kms",
					Usage: "add GCP KMS resource ID during rotation",
				},
				cli.StringFlag{
					Name:  "rm-gcp-kms",
					Usage: "remove GCP KMS resource ID during rotation",
				},
				cli.StringFlag{
					Name:  "add-azure-kv",
					Usage: "add Azure Key Vault URL during rotation",
				},
				cli.StringFlag{
					Name:  "rm-azure-kv",
					Usage: "remove Azure Key Vault URL during rotation",
				},
				cli.StringFlag{
					Name:  "add-hc-vault-transit",
					Usage: "add HashiCorp Vault Transit URI during rotation",
				},
				cli.StringFlag{
					Name:  "rm-hc-vault-transit",
					Usage: "remove HashiCorp Vault Transit URI during rotation",
				},
			},
			Action: handleRotate,
		},
		{
			Name:  "edit",
			Usage: "edit an encrypted file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "show-master-keys, s",
					Usage: "display master encryption keys in the file during editing",
				},
				cli.BoolFlag{
					Name:  "ignore-mac",
					Usage: "ignore Message Authentication Code during decryption",
				},
			},
			Action: handleEdit,
		},
		{
			Name:  "set",
			Usage: "set a specific key or branch in the input document",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
				cli.BoolFlag{
					Name:  "ignore-mac",
					Usage: "ignore Message Authentication Code during decryption",
				},
			},
			Action: handleSet,
		},
		{
			Name:  "unset",
			Usage: "unset a specific key or branch in the input document",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
				cli.BoolFlag{
					Name:  "ignore-mac",
					Usage: "ignore Message Authentication Code during decryption",
				},
			},
			Action: handleUnset,
		},
		{
			Name:  "updatekeys",
			Usage: "update the keys of SOPS files",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-type",
					Usage: "input format (json, yaml, dotenv, binary)",
				},
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
				cli.StringFlag{
					Name:  "add-age",
					Usage: "add age recipient to key groups",
				},
				cli.StringFlag{
					Name:  "rm-age",
					Usage: "remove age recipient from key groups",
				},
				cli.StringFlag{
					Name:  "add-pgp",
					Usage: "add PGP fingerprint to key groups",
				},
				cli.StringFlag{
					Name:  "rm-pgp",
					Usage: "remove PGP fingerprint from key groups",
				},
				cli.StringFlag{
					Name:  "add-kms",
					Usage: "add KMS ARN to key groups",
				},
				cli.StringFlag{
					Name:  "rm-kms",
					Usage: "remove KMS ARN from key groups",
				},
				cli.StringFlag{
					Name:  "add-gcp-kms",
					Usage: "add GCP KMS resource ID to key groups",
				},
				cli.StringFlag{
					Name:  "rm-gcp-kms",
					Usage: "remove GCP KMS resource ID from key groups",
				},
				cli.StringFlag{
					Name:  "add-azure-kv",
					Usage: "add Azure Key Vault URL to key groups",
				},
				cli.StringFlag{
					Name:  "rm-azure-kv",
					Usage: "remove Azure Key Vault URL from key groups",
				},
				cli.StringFlag{
					Name:  "add-hc-vault-transit",
					Usage: "add HashiCorp Vault Transit URI to key groups",
				},
				cli.StringFlag{
					Name:  "rm-hc-vault-transit",
					Usage: "remove HashiCorp Vault Transit URI from key groups",
				},
			},
			Action: handleUpdateKeys,
		},
		{
			Name:  "groups",
			Usage: "modify the groups on a SOPS file",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "add a new key group to a SOPS file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "input-type",
							Usage: "input format (json, yaml, dotenv, binary)",
						},
						cli.StringFlag{
							Name:  "output-type",
							Usage: "output format (json, yaml, dotenv, binary)",
						},
						cli.BoolFlag{
							Name:  "in-place, i",
							Usage: "write output back to the same file instead of stdout",
						},
						cli.StringFlag{
							Name:  "age, a",
							Usage: "age recipient for the new group",
						},
						cli.StringFlag{
							Name:  "pgp, p",
							Usage: "PGP fingerprint for the new group",
						},
						cli.StringFlag{
							Name:  "kms, k",
							Usage: "KMS ARN for the new group",
						},
						cli.StringFlag{
							Name:  "gcp-kms",
							Usage: "GCP KMS resource ID for the new group",
						},
						cli.StringFlag{
							Name:  "azure-kv",
							Usage: "Azure Key Vault URL for the new group",
						},
						cli.StringFlag{
							Name:  "hc-vault-transit",
							Usage: "HashiCorp Vault Transit URI for the new group",
						},
					},
					Action: handleGroupsAdd,
				},
				{
					Name:  "delete",
					Usage: "delete a key group from a SOPS file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "input-type",
							Usage: "input format (json, yaml, dotenv, binary)",
						},
						cli.StringFlag{
							Name:  "output-type",
							Usage: "output format (json, yaml, dotenv, binary)",
						},
						cli.BoolFlag{
							Name:  "in-place, i",
							Usage: "write output back to the same file instead of stdout",
						},
						cli.IntFlag{
							Name:  "group-index",
							Usage: "index of the group to delete (0-based)",
							Value: 0,
						},
					},
					Action: handleGroupsDelete,
				},
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
	inPlace := cliCtx.Bool("in-place")

	// Parse key groups from flags
	keyGroups, err := parseKeyGroups(cliCtx)
	if err != nil {
		return fmt.Errorf("failed to parse key groups: %w", err)
	}

	if len(keyGroups) == 0 {
		return ErrNoEncryptionKeys
	}

	if inPlace {
		// Encrypt and write back to the same file
		encrypted, err := operations.EncryptFile(inputFile, keyGroups, outputFormat)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}

		const fileMode = 0o600

		err = os.WriteFile(inputFile, encrypted, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Encrypt and output to stdout
		err = operations.EncryptFileToWriter(inputFile, keyGroups, outputFormat, os.Stdout)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}
	}

	return nil
}

func handleDecrypt(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")
	// Note: --extract and --ignore-mac flags are parsed but not yet fully implemented
	// They would require additional operations package support

	if inPlace {
		// Decrypt and write back to the same file
		decrypted, err := operations.DecryptFile(inputFile, outputFormat)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		const fileMode = 0o600

		err = os.WriteFile(inputFile, decrypted, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Decrypt and output to stdout
		err := operations.DecryptFileToWriter(inputFile, outputFormat, os.Stdout)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}
	}

	return nil
}

func handleRotate(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")

	if inPlace {
		// Rotate and write back to the same file
		rotated, err := operations.RotateFile(inputFile, outputFormat)
		if err != nil {
			return fmt.Errorf("rotation failed: %w", err)
		}

		const fileMode = 0o600

		err = os.WriteFile(inputFile, rotated, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Rotate and output to stdout
		err := operations.RotateFileToWriter(inputFile, outputFormat, os.Stdout)
		if err != nil {
			return fmt.Errorf("rotation failed: %w", err)
		}
	}

	return nil
}

func handleSet(cliCtx *cli.Context) error {
	const minSetArgs = 3
	if cliCtx.NArg() < minSetArgs {
		return ErrInvalidSetArgs
	}

	inputFile := cliCtx.Args().Get(0)
	key := cliCtx.Args().Get(1)

	const valueArgIdx = 2

	value := cliCtx.Args().Get(valueArgIdx)
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")

	// Parse tree path (simple implementation - just use the key as a single path element)
	treePath := []interface{}{key}

	if inPlace {
		// Set and write back to the same file
		modified, err := operations.SetValue(inputFile, treePath, value, outputFormat)
		if err != nil {
			return fmt.Errorf("set failed: %w", err)
		}

		const fileMode = 0o600

		err = os.WriteFile(inputFile, modified, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Set and output to stdout
		err := operations.SetValueToWriter(inputFile, treePath, value, outputFormat, os.Stdout)
		if err != nil {
			return fmt.Errorf("set failed: %w", err)
		}
	}

	return nil
}

func handleUnset(cliCtx *cli.Context) error {
	const minUnsetArgs = 2
	if cliCtx.NArg() < minUnsetArgs {
		return ErrInvalidUnsetArgs
	}

	inputFile := cliCtx.Args().Get(0)
	key := cliCtx.Args().Get(1)
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")

	// Parse tree path (simple implementation - just use the key as a single path element)
	treePath := []interface{}{key}

	if inPlace {
		// Unset and write back to the same file
		modified, err := operations.UnsetValue(inputFile, treePath, outputFormat)
		if err != nil {
			return fmt.Errorf("unset failed: %w", err)
		}

		const fileMode = 0o600

		err = os.WriteFile(inputFile, modified, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Unset and output to stdout
		err := operations.UnsetValueToWriter(inputFile, treePath, outputFormat, os.Stdout)
		if err != nil {
			return fmt.Errorf("unset failed: %w", err)
		}
	}

	return nil
}

//nolint:funlen // Edit command requires multiple steps for proper SOPS integration
func handleEdit(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")

	// Check for EDITOR environment variable
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return ErrNoEditor
	}

	// Decrypt file for editing
	plaintext, dataKey, err := operations.EditFile(inputFile, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to decrypt file for editing: %w", err)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "sops-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// Write plaintext to temp file
	_, err = tmpFile.Write(plaintext)
	if err != nil {
		_ = tmpFile.Close()

		return fmt.Errorf("failed to write temp file: %w", err)
	}

	_ = tmpFile.Close()

	// Launch editor
	//nolint:gosec,noctx // Editor and temp path are from user-controlled env var and temp file
	cmd := exec.Command(
		editor,
		tmpPath,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
	//nolint:gosec // Temp path is from secure temp file creation
	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Re-encrypt with original data key
	encrypted, err := operations.ReencryptFile(inputFile, editedContent, dataKey, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to re-encrypt file: %w", err)
	}

	// Write back to original file
	const fileMode = 0o600

	err = os.WriteFile(inputFile, encrypted, fileMode)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func handleUpdateKeys(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")

	// Parse new keys from flags
	newKeyGroups, err := parseKeyGroups(cliCtx)
	if err != nil {
		return fmt.Errorf("failed to parse key groups: %w", err)
	}

	if len(newKeyGroups) == 0 {
		return ErrNoEncryptionKeys
	}

	// Update keys
	output, err := operations.UpdateKeysFile(inputFile, newKeyGroups, outputFormat)
	if err != nil {
		return fmt.Errorf("updatekeys failed: %w", err)
	}

	if inPlace {
		const fileMode = 0o600

		err = os.WriteFile(inputFile, output, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		_, err = os.Stdout.Write(output)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

func handleGroupsAdd(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")

	// Parse new group from flags
	newKeyGroups, err := parseKeyGroups(cliCtx)
	if err != nil {
		return fmt.Errorf("failed to parse key group: %w", err)
	}

	if len(newKeyGroups) == 0 {
		return ErrNoEncryptionKeys
	}

	// Add the group
	output, err := operations.AddKeyGroup(inputFile, newKeyGroups[0], outputFormat)
	if err != nil {
		return fmt.Errorf("failed to add group: %w", err)
	}

	if inPlace {
		const fileMode = 0o600

		err = os.WriteFile(inputFile, output, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		_, err = os.Stdout.Write(output)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

func handleGroupsDelete(cliCtx *cli.Context) error {
	if cliCtx.NArg() < 1 {
		return ErrNoInputFile
	}

	inputFile := cliCtx.Args().First()
	outputFormat := cliCtx.String("output-type")
	inPlace := cliCtx.Bool("in-place")
	groupIndex := cliCtx.Int("group-index")

	if groupIndex < 0 {
		return ErrInvalidGroupIndex
	}

	// Delete the group
	output, err := operations.DeleteKeyGroup(inputFile, groupIndex, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if inPlace {
		const fileMode = 0o600

		err = os.WriteFile(inputFile, output, fileMode)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		_, err = os.Stdout.Write(output)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

//nolint:gocognit,cyclop,funlen // Key group parsing requires checking multiple flag combinations
func parseKeyGroups(cliCtx *cli.Context) ([]sops.KeyGroup, error) {
	var keyGroup sops.KeyGroup

	// Parse age recipients
	ageRecipients := cliCtx.String("age")
	if ageRecipients != "" {
		for _, recipient := range strings.Split(ageRecipients, ",") {
			recipient = strings.TrimSpace(recipient)
			if recipient != "" {
				masterKey := &age.MasterKey{
					Recipient: recipient,
				}
				keyGroup = append(keyGroup, masterKey)
			}
		}
	}

	// Parse PGP fingerprints
	pgpFingerprints := cliCtx.String("pgp")
	if pgpFingerprints != "" {
		for _, fingerprint := range strings.Split(pgpFingerprints, ",") {
			fingerprint = strings.TrimSpace(fingerprint)
			if fingerprint != "" {
				masterKey := pgp.NewMasterKeyFromFingerprint(fingerprint)
				keyGroup = append(keyGroup, masterKey)
			}
		}
	}

	// Parse KMS ARNs
	kmsARNs := cliCtx.String("kms")
	if kmsARNs != "" {
		for _, arn := range strings.Split(kmsARNs, ",") {
			arn = strings.TrimSpace(arn)
			if arn != "" {
				masterKey := &kms.MasterKey{
					Arn: arn,
				}
				keyGroup = append(keyGroup, masterKey)
			}
		}
	}

	// Parse GCP KMS resource IDs
	gcpKMSIDs := cliCtx.String("gcp-kms")
	if gcpKMSIDs != "" {
		for _, resourceID := range strings.Split(gcpKMSIDs, ",") {
			resourceID = strings.TrimSpace(resourceID)
			if resourceID != "" {
				masterKey := &gcpkms.MasterKey{
					ResourceID: resourceID,
				}
				keyGroup = append(keyGroup, masterKey)
			}
		}
	}

	// Parse Azure Key Vault URLs
	azureKVURLs := cliCtx.String("azure-kv")
	if azureKVURLs != "" {
		for _, vaultURL := range strings.Split(azureKVURLs, ",") {
			vaultURL = strings.TrimSpace(vaultURL)
			if vaultURL != "" {
				masterKey := &azkv.MasterKey{
					VaultURL: vaultURL,
					Name:     "", // Will be populated during encryption
					Version:  "", // Will be populated during encryption
				}
				keyGroup = append(keyGroup, masterKey)
			}
		}
	}

	// Parse HashiCorp Vault Transit URIs
	hcVaultURIs := cliCtx.String("hc-vault-transit")
	if hcVaultURIs != "" {
		for _, vaultURI := range strings.Split(hcVaultURIs, ",") {
			vaultURI = strings.TrimSpace(vaultURI)
			if vaultURI != "" {
				masterKey := &hcvault.MasterKey{
					VaultAddress: vaultURI,
					EnginePath:   "transit", // Default engine path
					KeyName:      "",        // Will be extracted from URI
				}
				keyGroup = append(keyGroup, masterKey)
			}
		}
	}

	if len(keyGroup) == 0 {
		return nil, nil
	}

	return []sops.KeyGroup{keyGroup}, nil
}
