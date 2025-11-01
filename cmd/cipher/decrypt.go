package cipher

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/getsops/sops/v3/aes"
	"github.com/spf13/cobra"
)

// NewDecryptCmd creates the decrypt subcommand.
func NewDecryptCmd() *cobra.Command {
	var keyFlag string

	var outputFlag string

	cmd := &cobra.Command{
		Use:   "decrypt <file>",
		Short: "Decrypt a file encrypted with AES-256-GCM",
		Long: `Decrypt a file that was encrypted using SOPS AES-256-GCM encryption.

The input file must contain a SOPS-format encrypted string:
  ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]

Examples:
  # Decrypt to stdout
  ksail cipher decrypt --key <base64-key> secrets.enc

  # Decrypt and save to a file
  ksail cipher decrypt --key <base64-key> --output secrets.txt secrets.enc`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDecryptRunE(cmd, args[0], keyFlag, outputFlag)
		},
	}

	cmd.Flags().StringVarP(&keyFlag, "key", "k", "", "Base64-encoded AES-256 key (32 bytes) used for encryption (required)")

	if err := cmd.MarkFlagRequired("key"); err != nil {
		// This should never happen with a valid flag name, but handle it gracefully
		panic(fmt.Sprintf("failed to mark key flag as required: %v", err))
	}

	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path. If not provided, prints to stdout.")

	return cmd
}

func handleDecryptRunE(cmd *cobra.Command, inputFile, keyFlag, outputFlag string) error {
	tmr := timer.New()
	tmr.Start()

	// Read the encrypted file
	encryptedData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Decode the key
	key, err := base64DecodeKey(keyFlag)
	if err != nil {
		return err
	}

	// Decrypt the data
	cipher := aes.NewCipher()

	decrypted, err := cipher.Decrypt(string(encryptedData), key, "")
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Convert decrypted interface to string
	decryptedStr, ok := decrypted.(string)
	if !ok {
		return fmt.Errorf("unexpected decrypted data type: %T", decrypted)
	}

	// Write output
	if outputFlag != "" {
		err = os.WriteFile(outputFlag, []byte(decryptedStr), filePermOwner)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		total, stage := tmr.GetTiming()
		timingStr := notify.FormatTiming(total, stage, false)
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "decrypted file written to " + outputFlag + " " + timingStr,
			Writer:  cmd.OutOrStdout(),
		})
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), decryptedStr)

		total, stage := tmr.GetTiming()
		timingStr := notify.FormatTiming(total, stage, false)
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "decryption complete " + timingStr,
			Writer:  cmd.OutOrStdout(),
		})
	}

	return nil
}

// base64DecodeKey decodes and validates a base64-encoded AES-256 key.
func base64DecodeKey(keyFlag string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}
	if len(key) != aesKeySize {
		return nil, fmt.Errorf("key must be %d bytes for AES-256, got %d bytes", aesKeySize, len(key))
	}

	return key, nil
}
