// Package main implements the JSON schema generator for KSail configuration.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/invopop/jsonschema"
)

const (
	dirPermissions  = 0o750 // Directory permissions (rwxr-x---)
	filePermissions = 0o600 // File permissions (rw-------)
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Generate JSON schema from the Cluster type
	reflector := jsonschema.Reflector{
		DoNotReference: true,
	}
	schema := reflector.Reflect(&v1alpha1.Cluster{})

	// Set the schema ID and title
	schema.ID = "https://ksail.dev/schemas/ksail.json"
	schema.Title = "KSail Cluster Configuration"
	schema.Description = "Schema for KSail cluster configuration (ksail.yaml)"

	// Marshal to JSON with indentation
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling schema: %w", err)
	}

	// Determine output path - default to schemas/ksail.json
	outputPath := "schemas/ksail.json"
	if len(args) > 0 {
		outputPath = args[0]
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)

	err = os.MkdirAll(dir, dirPermissions)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	// Write the schema to the file
	err = os.WriteFile(outputPath, schemaJSON, filePermissions)
	if err != nil {
		return fmt.Errorf("error writing schema to %s: %w", outputPath, err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "JSON schema generated successfully: %s\n", outputPath)

	return nil
}
