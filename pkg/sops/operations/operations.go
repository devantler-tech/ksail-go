// Package operations provides direct SOPS operations using Go libraries.
package operations

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/getsops/sops/v3"                 //nolint:depguard // Required for SOPS operations
	"github.com/getsops/sops/v3/aes"             //nolint:depguard // Required for AES encryption
	"github.com/getsops/sops/v3/cmd/sops/common" //nolint:depguard // Required for SOPS operations
	"github.com/getsops/sops/v3/keyservice"      //nolint:depguard // Required for key services
)

var (
	// ErrEmptyFile is returned when a file contains no documents.
	ErrEmptyFile = errors.New("file must contain at least one document")
	// ErrGenerateDataKey is returned when data key generation fails.
	ErrGenerateDataKey = errors.New("failed to generate data key")
)

// DecryptFile decrypts a SOPS-encrypted file and returns the plaintext.
func DecryptFile(inputPath string, outputFormat string) ([]byte, error) {
	// Create cipher
	cipher := aes.NewCipher()

	// Determine format and create store
	store := common.DefaultStoreForPathOrFormat(nil, inputPath, outputFormat)

	// Create key services
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}

	// Load encrypted file
	tree, err := common.LoadEncryptedFile(store, inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load encrypted file: %w", err)
	}

	// Decrypt the tree
	_, err = common.DecryptTree(common.DecryptTreeOpts{
		Cipher:      cipher,
		IgnoreMac:   false,
		Tree:        tree,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tree: %w", err)
	}

	// Emit plaintext
	plaintext, err := store.EmitPlainFile(tree.Branches)
	if err != nil {
		return nil, fmt.Errorf("failed to emit plaintext: %w", err)
	}

	return plaintext, nil
}

// EncryptFile encrypts a plaintext file using SOPS.
func EncryptFile(inputPath string, keyGroups []sops.KeyGroup, outputFormat string) ([]byte, error) {
	// Create cipher
	cipher := aes.NewCipher()

	// Determine format and create stores
	inputStore := common.DefaultStoreForPathOrFormat(nil, inputPath, outputFormat)
	outputStore := inputStore

	// Create key services
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}

	// Load plaintext file
	fileBytes, err := os.ReadFile(inputPath) //nolint:gosec // Input path is provided by user
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	branches, err := inputStore.LoadPlainFile(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load plaintext file: %w", err)
	}

	if len(branches) < 1 {
		return nil, ErrEmptyFile
	}

	// Create tree
	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups: keyGroups,
		},
		FilePath: inputPath,
	}

	// Generate data key
	dataKey, errs := tree.GenerateDataKeyWithKeyServices(keyServices)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrGenerateDataKey, errs)
	}

	// Encrypt tree
	err = common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    &tree,
		Cipher:  cipher,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tree: %w", err)
	}

	// Emit encrypted file
	encryptedFile, err := outputStore.EmitEncryptedFile(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to emit encrypted file: %w", err)
	}

	return encryptedFile, nil
}

// DecryptFileToWriter decrypts a file and writes to the provided writer.
func DecryptFileToWriter(inputPath string, outputFormat string, writer io.Writer) error {
	plaintext, err := DecryptFile(inputPath, outputFormat)
	if err != nil {
		return err
	}

	_, writeErr := writer.Write(plaintext)
	if writeErr != nil {
		return fmt.Errorf("failed to write plaintext: %w", writeErr)
	}

	return nil
}

// EncryptFileToWriter encrypts a file and writes to the provided writer.
func EncryptFileToWriter(
	inputPath string, keyGroups []sops.KeyGroup, outputFormat string, writer io.Writer,
) error {
	encrypted, err := EncryptFile(inputPath, keyGroups, outputFormat)
	if err != nil {
		return err
	}

	_, writeErr := writer.Write(encrypted)
	if writeErr != nil {
		return fmt.Errorf("failed to write encrypted data: %w", writeErr)
	}

	return nil
}

// RotateFile rotates the data encryption key of a SOPS file.
func RotateFile(inputPath string, outputFormat string) ([]byte, error) {
	// Create cipher
	cipher := aes.NewCipher()

	// Determine format and create stores
	store := common.DefaultStoreForPathOrFormat(nil, inputPath, outputFormat)

	// Create key services
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}

	// Load encrypted file
	tree, err := common.LoadEncryptedFileWithBugFixes(common.GenericDecryptOpts{
		Cipher:      cipher,
		InputStore:  store,
		InputPath:   inputPath,
		IgnoreMAC:   false,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load encrypted file: %w", err)
	}

	// Decrypt the file
	_, err = common.DecryptTree(common.DecryptTreeOpts{
		Cipher:      cipher,
		IgnoreMac:   false,
		Tree:        tree,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tree: %w", err)
	}

	// Create a new data key
	dataKey, errs := tree.GenerateDataKeyWithKeyServices(keyServices)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrGenerateDataKey, errs)
	}

	// Reencrypt the file with the new key
	err = common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    tree,
		Cipher:  cipher,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tree: %w", err)
	}

	// Emit encrypted file
	encryptedFile, err := store.EmitEncryptedFile(*tree)
	if err != nil {
		return nil, fmt.Errorf("failed to emit encrypted file: %w", err)
	}

	return encryptedFile, nil
}

// SetValue sets a value at the specified path in a SOPS file.
func SetValue(
	inputPath string, treePath []interface{}, value interface{}, outputFormat string,
) ([]byte, error) {
	// Create cipher
	cipher := aes.NewCipher()

	// Determine format and create stores
	store := common.DefaultStoreForPathOrFormat(nil, inputPath, outputFormat)

	// Create key services
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}

	// Load encrypted file
	tree, err := common.LoadEncryptedFileWithBugFixes(common.GenericDecryptOpts{
		Cipher:      cipher,
		InputStore:  store,
		InputPath:   inputPath,
		IgnoreMAC:   false,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load encrypted file: %w", err)
	}

	// Decrypt the file
	dataKey, err := common.DecryptTree(common.DecryptTreeOpts{
		Cipher:      cipher,
		IgnoreMac:   false,
		Tree:        tree,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tree: %w", err)
	}

	// Set the value
	tree.Branches[0], _ = tree.Branches[0].Set(treePath, value)

	// Re-encrypt
	err = common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    tree,
		Cipher:  cipher,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tree: %w", err)
	}

	// Emit encrypted file
	encryptedFile, err := store.EmitEncryptedFile(*tree)
	if err != nil {
		return nil, fmt.Errorf("failed to emit encrypted file: %w", err)
	}

	return encryptedFile, nil
}

// UnsetValue unsets a value at the specified path in a SOPS file.
func UnsetValue(inputPath string, treePath []interface{}, outputFormat string) ([]byte, error) {
	// Create cipher
	cipher := aes.NewCipher()

	// Determine format and create stores
	store := common.DefaultStoreForPathOrFormat(nil, inputPath, outputFormat)

	// Create key services
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}

	// Load encrypted file
	tree, err := common.LoadEncryptedFileWithBugFixes(common.GenericDecryptOpts{
		Cipher:      cipher,
		InputStore:  store,
		InputPath:   inputPath,
		IgnoreMAC:   false,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load encrypted file: %w", err)
	}

	// Decrypt the file
	dataKey, err := common.DecryptTree(common.DecryptTreeOpts{
		Cipher:      cipher,
		IgnoreMac:   false,
		Tree:        tree,
		KeyServices: keyServices,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tree: %w", err)
	}

	// Unset the value
	newBranch, err := tree.Branches[0].Unset(treePath)
	if err != nil {
		return nil, fmt.Errorf("failed to unset value: %w", err)
	}
	tree.Branches[0] = newBranch

	// Re-encrypt
	err = common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    tree,
		Cipher:  cipher,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tree: %w", err)
	}

	// Emit encrypted file
	encryptedFile, err := store.EmitEncryptedFile(*tree)
	if err != nil {
		return nil, fmt.Errorf("failed to emit encrypted file: %w", err)
	}

	return encryptedFile, nil
}

// RotateFileToWriter rotates a file's encryption key and writes to the provided writer.
func RotateFileToWriter(inputPath string, outputFormat string, writer io.Writer) error {
	rotated, err := RotateFile(inputPath, outputFormat)
	if err != nil {
		return err
	}

	_, writeErr := writer.Write(rotated)
	if writeErr != nil {
		return fmt.Errorf("failed to write rotated file: %w", writeErr)
	}

	return nil
}

// SetValueToWriter sets a value and writes to the provided writer.
func SetValueToWriter(
	inputPath string, treePath []interface{}, value interface{}, outputFormat string, writer io.Writer,
) error {
	modified, err := SetValue(inputPath, treePath, value, outputFormat)
	if err != nil {
		return err
	}

	_, writeErr := writer.Write(modified)
	if writeErr != nil {
		return fmt.Errorf("failed to write modified file: %w", writeErr)
	}

	return nil
}

// UnsetValueToWriter unsets a value and writes to the provided writer.
func UnsetValueToWriter(inputPath string, treePath []interface{}, outputFormat string, writer io.Writer) error {
	modified, err := UnsetValue(inputPath, treePath, outputFormat)
	if err != nil {
		return err
	}

	_, writeErr := writer.Write(modified)
	if writeErr != nil {
		return fmt.Errorf("failed to write modified file: %w", writeErr)
	}

	return nil
}
