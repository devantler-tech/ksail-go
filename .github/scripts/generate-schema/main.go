package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/invopop/jsonschema"
)

func main() {
	// Generate JSON schema from the Cluster type
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	
	schema := reflector.Reflect(&v1alpha1.Cluster{})
	
	// Add schema metadata
	schema.Title = "KSail Cluster Configuration"
	schema.Description = "JSON schema for KSail cluster configuration (ksail.yaml)"
	
	// Marshal to JSON with pretty printing
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}
	
	// Determine output path
	outputPath := "schemas/ksail-config.schema.json"
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
		os.Exit(1)
	}
	
	// Write schema to file
	if err := os.WriteFile(outputPath, schemaJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema to %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	
	fmt.Printf("Successfully generated JSON schema at %s\n", outputPath)
}
