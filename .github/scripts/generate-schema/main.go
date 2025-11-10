// Package main provides a CLI tool to generate JSON schema from KSail config types.
package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// AllowAdditionalProperties: false - Enforce strict validation, reject unknown fields
	// DoNotReference: true - Inline all type definitions for simpler schema structure
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
		Mapper:                    customTypeMapper,
	}

	schema := reflector.Reflect(&v1alpha1.Cluster{})

	// Customize schema metadata and properties
	customizeSchema(schema)

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

// customizeSchema customizes the generated schema with metadata and property constraints.
func customizeSchema(schema *jsonschema.Schema) {
	// Add schema metadata
	schema.ID = ""
	schema.Title = "KSail Cluster Configuration"
	schema.Description = "JSON schema for KSail cluster configuration (ksail.yaml)"

	// Make only spec required at the root level based on omitzero tags
	schema.Required = []string{"spec"}

	if schema.Properties == nil {
		return
	}

	customizeRootProperties(schema.Properties)
	customizeSpecProperties(schema.Properties)
}

// customizeRootProperties adds enum constraints to root-level properties.
func customizeRootProperties(properties *orderedmap.OrderedMap[string, *jsonschema.Schema]) {
	// Add enum constraint for kind
	if kindProp, ok := properties.Get("kind"); ok && kindProp != nil {
		kindProp.Enum = []any{"Cluster"}
	}

	// Add enum constraint for apiVersion
	if apiVersionProp, ok := properties.Get("apiVersion"); ok && apiVersionProp != nil {
		apiVersionProp.Enum = []any{"ksail.dev/v1alpha1"}
	}
}

// customizeSpecProperties fixes required fields for spec and nested properties.
func customizeSpecProperties(properties *orderedmap.OrderedMap[string, *jsonschema.Schema]) {
	specProp, ok := properties.Get("spec")
	if !ok || specProp == nil || specProp.Properties == nil {
		return
	}

	// Fix required fields for spec - all fields have omitzero so they're optional
	specProp.Required = nil

	// Also fix required fields for connection
	if connProp, ok := specProp.Properties.Get("connection"); ok && connProp != nil {
		connProp.Required = nil
	}

	// Also fix required fields for options (all fields have omitzero so they're optional)
	if optionsProp, ok := specProp.Properties.Get("options"); ok && optionsProp != nil {
		optionsProp.Required = nil
	}
}

// customTypeMapper provides custom schema mappings for specific types.
func customTypeMapper(reflectType reflect.Type) *jsonschema.Schema {
	// Handle metav1.Duration - it marshals as a string like "5m", "1h"
	if reflectType == reflect.TypeFor[metav1.Duration]() {
		return &jsonschema.Schema{
			Type:    "string",
			Pattern: "^[0-9]+(ns|us|µs|ms|s|m|h)$",
		}
	}

	// Handle Distribution enum
	if reflectType == reflect.TypeFor[v1alpha1.Distribution]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"Kind", "K3d"},
		}
	}

	// Handle CNI enum
	if reflectType == reflect.TypeFor[v1alpha1.CNI]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"Default", "Cilium"},
		}
	}

	// Handle CSI enum
	if reflectType == reflect.TypeFor[v1alpha1.CSI]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"Default", "LocalPathStorage"},
		}
	}

	// Handle IngressController enum
	if reflectType == reflect.TypeFor[v1alpha1.IngressController]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"Default", "Traefik", "None"},
		}
	}

	// Handle GatewayController enum
	if reflectType == reflect.TypeFor[v1alpha1.GatewayController]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"Default", "Traefik", "Cilium", "None"},
		}
	}

	// Handle MetricsServer enum
	if reflectType == reflect.TypeFor[v1alpha1.MetricsServer]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"Enabled", "Disabled"},
		}
	}

	// Handle GitOpsEngine enum
	if reflectType == reflect.TypeFor[v1alpha1.GitOpsEngine]() {
		return &jsonschema.Schema{
			Type: "string",
			Enum: []any{"None"},
		}
	}

	// Return nil to use default mapping for other types
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
