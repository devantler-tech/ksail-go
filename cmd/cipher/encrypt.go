package cipher

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/getsops/sops/v3/aes"
	"github.com/spf13/cobra"
)

// NewEncryptCmd creates the encrypt subcommand.
func NewEncryptCmd() *cobra.Command {
	var keyFlag string

	var outputFlag string

	cmd := &cobra.Command{
		Use:   "encrypt <file>",
		Short: "Encrypt a file using AES-256-GCM",
		Long: `Encrypt a file using AES-256-GCM encryption from SOPS.

The encrypted output will be in SOPS format:
  ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]

Examples:
  # Encrypt with a random key (key will be displayed)
  ksail cipher encrypt secrets.txt

  # Encrypt with a specific base64-encoded key
  ksail cipher encrypt --key <base64-key> secrets.txt

  # Encrypt and save to a different file
  ksail cipher encrypt --output secrets.enc secrets.txt`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleEncryptRunE(cmd, args[0], keyFlag, outputFlag)
		},
	}

	cmd.Flags().StringVarP(&keyFlag, "key", "k", "", "Base64-encoded AES-256 key (32 bytes). If not provided, a random key will be generated.")
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path. If not provided, prints to stdout.")

	return cmd
}

func handleEncryptRunE(cmd *cobra.Command, inputFile, keyFlag, outputFlag string) error {
	tmr := timer.New()
	tmr.Start()

	// Read the input file
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Get or generate the key
	var key []byte
	if keyFlag != "" {
		key, err = base64.StdEncoding.DecodeString(keyFlag)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}

		if len(key) != aesKeySize {
			return fmt.Errorf("key must be %d bytes for AES-256, got %d bytes", aesKeySize, len(key))
		}
	} else {
		// Generate a random 32-byte key for AES-256
		key = make([]byte, aesKeySize)

		_, err = rand.Read(key)
		if err != nil {
			return fmt.Errorf("failed to generate random key: %w", err)
		}
		// Display the generated key so the user can use it for decryption
		encodedKey := base64.StdEncoding.EncodeToString(key)
		notify.WriteMessage(notify.Message{
			Type:    notify.InfoType,
			Content: "Generated key: " + encodedKey,
			Writer:  cmd.OutOrStdout(),
		})
		notify.WriteMessage(notify.Message{
			Type:    notify.InfoType,
			Content: "Save this key securely - you will need it to decrypt the file",
			Writer:  cmd.OutOrStdout(),
		})
	}

	// Encrypt the data
	cipher := aes.NewCipher()

	encrypted, err := cipher.Encrypt(string(plaintext), key, "")
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Write output
	if outputFlag != "" {
		err = os.WriteFile(outputFlag, []byte(encrypted), filePermOwner)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		total, stage := tmr.GetTiming()
		timingStr := notify.FormatTiming(total, stage, false)
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "encrypted file written to " + outputFlag + " " + timingStr,
			Writer:  cmd.OutOrStdout(),
		})
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), encrypted)

		total, stage := tmr.GetTiming()
		timingStr := notify.FormatTiming(total, stage, false)
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "encryption complete " + timingStr,
			Writer:  cmd.OutOrStdout(),
		})
	}

	return nil
}
