// Package main provides a CLI tool to generate JSON schema from KSail config types.
package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/invopop/jsonschema"
)

const (
	dirPermissions  = 0o750
	filePermissions = 0o600
)

func main() {
	err := run(os.Stdout, os.Stderr, os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func run(stdout, stderr io.Writer, args []string) error {
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
		_, _ = io.WriteString(stderr, "Error marshaling schema: "+err.Error()+"\n")

		return wrapError("marshal schema", err)
	}

	// Determine output path
	outputPath := "schemas/ksail-config.schema.json"
	if len(args) > 1 {
		outputPath = args[1]
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)

	err = os.MkdirAll(dir, dirPermissions)
	if err != nil {
		_, _ = io.WriteString(stderr, "Error creating directory "+dir+": "+err.Error()+"\n")

		return wrapError("create directory", err)
	}

	// Write schema to file
	err = os.WriteFile(outputPath, schemaJSON, filePermissions)
	if err != nil {
		_, _ = io.WriteString(stderr, "Error writing schema to "+outputPath+": "+err.Error()+"\n")

		return wrapError("write schema file", err)
	}

	_, _ = io.WriteString(stdout, "Successfully generated JSON schema at "+outputPath+"\n")

	return nil
}

func wrapError(msg string, err error) error {
	if err == nil {
		return nil
	}

	return &schemaError{
		msg: msg,
		err: err,
	}
}

type schemaError struct {
	msg string
	err error
}

func (e *schemaError) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e *schemaError) Unwrap() error {
	return e.err
}
