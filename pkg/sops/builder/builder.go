// Package builder provides a builder for creating SOPS urfave/cli applications.
package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/devantler-tech/ksail-go/pkg/sops/operations"
	"github.com/getsops/sops/v3"         //nolint:depguard // Required for SOPS operations
	"github.com/getsops/sops/v3/age"     //nolint:depguard // Required for age encryption
	"github.com/getsops/sops/v3/pgp"     //nolint:depguard // Required for PGP encryption
	"github.com/getsops/sops/v3/version" //nolint:depguard // Required for sops version info
	"github.com/urfave/cli"              //nolint:depguard // This package wraps urfave/cli apps
)

var (
	// ErrNotImplemented is returned when a command is not yet implemented.
	ErrNotImplemented   = errors.New("command not yet implemented with Go libraries")
	// ErrNoInputFile is returned when no input file is specified.
	ErrNoInputFile      = errors.New("no input file specified")
	// ErrNoEncryptionKeys is returned when no encryption keys are specified.
	ErrNoEncryptionKeys = errors.New("no encryption keys specified (use --age or --pgp)")
	// ErrInvalidSetArgs is returned when set command has invalid arguments.
	ErrInvalidSetArgs   = errors.New("usage: set <file> <key> <value>")
	// ErrInvalidUnsetArgs is returned when unset command has invalid arguments.
	ErrInvalidUnsetArgs = errors.New("usage: unset <file> <key>")
	// ErrNoEditor is returned when EDITOR environment variable is not set.
	ErrNoEditor         = errors.New("EDITOR environment variable not set")
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
			},
			Action: handleRotate,
		},
		{
			Name:  "edit",
			Usage: "edit an encrypted file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
			},
			Action: handleEdit,
		},
		{
			Name:  "set",
			Usage: "set a specific key or branch in the input document",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
			},
			Action: handleSet,
		},
		{
			Name:  "unset",
			Usage: "unset a specific key or branch in the input document",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output-type",
					Usage: "output format (json, yaml, dotenv, binary)",
				},
				cli.BoolFlag{
					Name:  "in-place, i",
					Usage: "write output back to the same file instead of stdout",
				},
			},
			Action: handleUnset,
		},
		{
			Name:  "updatekeys",
			Usage: "update the keys of SOPS files",
			Flags: []cli.Flag{
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
					Name:  "add-pgp",
					Usage: "add PGP fingerprint to key groups",
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
							Name:  "output-type",
							Usage: "output format (json, yaml, dotenv, binary)",
						},
						cli.BoolFlag{
							Name:  "in-place, i",
							Usage: "write output back to the same file instead of stdout",
						},
						cli.StringFlag{
							Name:  "age",
							Usage: "age recipient for the new group",
						},
						cli.StringFlag{
							Name:  "pgp",
							Usage: "PGP fingerprint for the new group",
						},
					},
					Action: handleGroupsAdd,
				},
				{
					Name:  "delete",
					Usage: "delete a key group from a SOPS file",
					Flags: []cli.Flag{
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

		const fileMode = 0600
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

		const fileMode = 0600
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

		const fileMode = 0600
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
	cmd := exec.Command(editor, tmpPath) //nolint:gosec // Editor is from EDITOR env var, user controlled
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
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
	const fileMode = 0600
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
		const fileMode = 0600
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
		const fileMode = 0600
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
		const fileMode = 0600
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
