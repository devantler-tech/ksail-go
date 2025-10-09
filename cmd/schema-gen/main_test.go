package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGeneratesSchema(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-schema.json")

	// Run with custom output path
	err := run([]string{outputPath})
	require.NoError(t, err)

	// Verify the schema file was created
	require.FileExists(t, outputPath)

	// Verify the schema is valid JSON
	schemaBytes, err := os.ReadFile(outputPath) //nolint:gosec // test file path
	require.NoError(t, err)

	var schema map[string]interface{}

	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err)

	// Verify key schema properties
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema["$schema"])
	assert.Equal(t, "https://ksail.dev/schemas/ksail-cluster-schema.json", schema["$id"])
	assert.Equal(t, "KSail Cluster Configuration", schema["title"])
	assert.Equal(t, "Schema for KSail cluster configuration (ksail.yaml)", schema["description"])

	// Verify properties exist
	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "properties should be a map")
	assert.Contains(t, props, "kind")
	assert.Contains(t, props, "apiVersion")
	assert.Contains(t, props, "spec")
}

//nolint:paralleltest // Cannot use t.Parallel with t.Chdir
func TestRunGeneratesSchemaWithDefaultPath(t *testing.T) {
	// Create a temporary directory and change to it
	tempDir := t.TempDir()

	t.Chdir(tempDir)

	// Run with no arguments (use default path)
	err := run([]string{})
	require.NoError(t, err)

	// Verify the schema file was created at default location
	defaultPath := filepath.Join(tempDir, "schemas", "ksail-cluster-schema.json")
	require.FileExists(t, defaultPath)

	// Verify the schema is valid JSON
	schemaBytes, err := os.ReadFile(defaultPath) //nolint:gosec // test file path
	require.NoError(t, err)

	var schema map[string]interface{}

	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err)

	// Verify it has the expected structure
	assert.NotEmpty(t, schema["$schema"])
	assert.NotEmpty(t, schema["$id"])
}

func TestRunOverwritesExistingFile(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing-schema.json")

	// Create an existing file with some content
	existingContent := []byte(`{"old": "content"}`)

	err := os.WriteFile(outputPath, existingContent, 0o600)
	require.NoError(t, err)

	// Run with the output path
	err = run([]string{outputPath})
	require.NoError(t, err)

	// Verify the file was overwritten
	newContent, err := os.ReadFile(outputPath) //nolint:gosec // test file path
	require.NoError(t, err)
	assert.NotEqual(t, string(existingContent), string(newContent))

	// Verify it's a valid schema now
	var schema map[string]interface{}

	err = json.Unmarshal(newContent, &schema)
	require.NoError(t, err)
	assert.Equal(t, "KSail Cluster Configuration", schema["title"])
}

func TestRunReturnsErrorOnInvalidDirectory(t *testing.T) {
	t.Parallel()

	// Try to write to a directory that can't be created (invalid path)
	err := run([]string{"/\x00invalid/path/schema.json"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error creating directory")
}
