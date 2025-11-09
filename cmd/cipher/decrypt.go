package cipher

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/cmd/sops/codes"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/keyservice"
	"github.com/getsops/sops/v3/stores/json"
	"github.com/spf13/cobra"
)

const notBinaryHint = "This is likely not an encrypted binary file? " +
	"If not, use --output-type to select the correct output type."

// decryptOpts contains all options needed for the decryption operation.
type decryptOpts struct {
	Cipher          sops.Cipher
	InputStore      sops.Store
	OutputStore     sops.Store
	InputPath       string
	OutputPath      string
	ReadFromStdin   bool
	IgnoreMAC       bool
	Extract         []interface{}
	KeyServices     []keyservice.KeyServiceClient
	DecryptionOrder []string
}

// decryptTree loads and decrypts a SOPS tree from the input file.
// It handles loading the encrypted file and decrypting its contents.
func decryptTree(opts decryptOpts) (*sops.Tree, error) {
	tree, err := common.LoadEncryptedFileWithBugFixes(common.GenericDecryptOpts{
		Cipher:        opts.Cipher,
		InputStore:    opts.InputStore,
		InputPath:     opts.InputPath,
		ReadFromStdin: opts.ReadFromStdin,
		IgnoreMAC:     opts.IgnoreMAC,
		KeyServices:   opts.KeyServices,
	})
	if err != nil {
		return nil, err
	}

	_, err = common.DecryptTree(common.DecryptTreeOpts{
		Cipher:          opts.Cipher,
		IgnoreMac:       opts.IgnoreMAC,
		Tree:            tree,
		KeyServices:     opts.KeyServices,
		DecryptionOrder: opts.DecryptionOrder,
	})
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// decrypt performs the core decryption logic for a file.
// It loads the encrypted file, decrypts it, and handles extraction if specified.
func decrypt(opts decryptOpts) ([]byte, error) {
	tree, err := decryptTree(opts)
	if err != nil {
		return nil, err
	}

	if len(opts.Extract) > 0 {
		return extract(tree, opts.Extract, opts.OutputStore)
	}

	decryptedFile, err := opts.OutputStore.EmitPlainFile(tree.Branches)
	if errors.Is(err, json.BinaryStoreEmitPlainError) {
		err = fmt.Errorf("%s\n\n%s", err.Error(), notBinaryHint)
	}
	if err != nil {
		return nil, common.NewExitError(
			fmt.Sprintf("Error dumping file: %s", err),
			codes.ErrorDumpingTree,
		)
	}

	return decryptedFile, nil
}

// extract retrieves a specific value or subtree from the decrypted tree.
// It supports extracting nested keys using a path array.
func extract(tree *sops.Tree, path []interface{}, outputStore sops.Store) ([]byte, error) {
	v, err := tree.Branches[0].Truncate(path)
	if err != nil {
		return nil, fmt.Errorf("error truncating tree: %w", err)
	}

	if newBranch, ok := v.(sops.TreeBranch); ok {
		tree.Branches[0] = newBranch
		decrypted, err := outputStore.EmitPlainFile(tree.Branches)
		if errors.Is(err, json.BinaryStoreEmitPlainError) {
			err = fmt.Errorf("%s\n\n%s", err.Error(), notBinaryHint)
		}
		if err != nil {
			return nil, common.NewExitError(
				fmt.Sprintf("Error dumping file: %s", err),
				codes.ErrorDumpingTree,
			)
		}
		return decrypted, nil
	} else if str, ok := v.(string); ok {
		return []byte(str), nil
	}

	bytes, err := outputStore.EmitValue(v)
	if err != nil {
		return nil, common.NewExitError(
			fmt.Sprintf("Error dumping tree: %s", err),
			codes.ErrorDumpingTree,
		)
	}

	return bytes, nil
}

// NewDecryptCmd creates and returns the decrypt command.
func NewDecryptCmd() *cobra.Command {
	var (
		extract   string
		ignoreMac bool
		output    string
	)

	cmd := &cobra.Command{
		Use:   "decrypt <file>",
		Short: "Decrypt a file with SOPS",
		Long: `Decrypt a file using SOPS (Secrets OPerationS).

SOPS supports multiple key management systems:
  - age recipients
  - PGP fingerprints
  - AWS KMS
  - GCP KMS
  - Azure Key Vault
  - HashiCorp Vault

Example:
  ksail cipher decrypt secrets.yaml
  ksail cipher decrypt secrets.yaml --extract '["data"]["password"]'
  ksail cipher decrypt secrets.yaml --output plaintext.yaml
  ksail cipher decrypt secrets.yaml --ignore-mac`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDecryptRunE(cmd, args, extract, ignoreMac, output)
		},
	}

	cmd.Flags().StringVarP(&extract, "extract", "e", "", "extract a specific key from the decrypted file (JSONPath format)")
	cmd.Flags().BoolVar(&ignoreMac, "ignore-mac", false, "ignore Message Authentication Code (MAC) check")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: stdout)")

	return cmd
}

const decryptedFilePermissions = 0o600

// handleDecryptRunE is the main handler for the decrypt command.
// It orchestrates the decryption workflow: determining file stores,
// setting up decryption options, decrypting the file, and writing
// the decrypted content to stdout or a file.
func handleDecryptRunE(cmd *cobra.Command, args []string, extract string, ignoreMac bool, output string) error {
	var inputPath string
	readFromStdin := len(args) == 0

	if !readFromStdin {
		inputPath = args[0]
	}

	inputStore, outputStore, err := getDecryptStores(inputPath, readFromStdin)
	if err != nil {
		return err
	}

	var extractPath []interface{}
	if extract != "" {
		extractPath, err = parseExtractPath(extract)
		if err != nil {
			return fmt.Errorf("failed to parse extract path: %w", err)
		}
	}

	opts := decryptOpts{
		Cipher:          aes.NewCipher(),
		InputStore:      inputStore,
		OutputStore:     outputStore,
		InputPath:       inputPath,
		ReadFromStdin:   readFromStdin,
		IgnoreMAC:       ignoreMac,
		Extract:         extractPath,
		KeyServices:     []keyservice.KeyServiceClient{keyservice.NewLocalClient()},
		DecryptionOrder: []string{},
	}

	decryptedData, err := decrypt(opts)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write to output file if specified, otherwise write to stdout
	if output != "" {
		err = os.WriteFile(output, decryptedData, decryptedFilePermissions)
		if err != nil {
			return fmt.Errorf("failed to write decrypted file: %w", err)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Successfully decrypted to %s\n", output)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		_, err = cmd.OutOrStdout().Write(decryptedData)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

// getDecryptStores returns the appropriate SOPS stores for decryption.
// When reading from stdin, it defaults to YAML format.
// For JSON format from stdin, users can pipe to a file first.
func getDecryptStores(inputPath string, readFromStdin bool) (sops.Store, sops.Store, error) {
	if readFromStdin {
		// Default to YAML for stdin - most common format
		return getStores("stdin.yaml")
	}

	return getStores(inputPath)
}

// parseExtractPath converts a JSONPath-like extract string into a path array.
// Example: '["data"]["password"]' -> []interface{}{"data", "password"}
func parseExtractPath(extract string) ([]interface{}, error) {
	// Remove outer quotes if present
	extract = strings.Trim(extract, "'\"")

	// Parse the JSONPath format: ["key1"]["key2"]
	var path []interface{}
	parts := strings.Split(extract, "][")

	for _, part := range parts {
		// Clean up the part
		part = strings.TrimPrefix(part, "[")
		part = strings.TrimSuffix(part, "]")
		part = strings.Trim(part, "\"'")

		if part != "" {
			path = append(path, part)
		}
	}

	if len(path) == 0 {
		return nil, errors.New("invalid extract path format")
	}

	return path, nil
}
