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
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	// Determine output path - default to schemas/ksail.json
	outputPath := "schemas/ksail.json"
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)

	err = os.MkdirAll(dir, dirPermissions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	// Write the schema to the file
	err = os.WriteFile(outputPath, schemaJSON, filePermissions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema to %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "JSON schema generated successfully: %s\n", outputPath)
}
