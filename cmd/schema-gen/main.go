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

// addPropertyDescriptions adds human-readable descriptions to schema properties.
// Descriptions are based on CLI flag descriptions used in the config manager.
func addPropertyDescriptions(schema *jsonschema.Schema) {
	// Add descriptions to top-level properties
	if schema.Properties != nil {
		if apiVersionProp, ok := schema.Properties.Get("apiVersion"); ok && apiVersionProp != nil {
			apiVersionProp.Description = "API version of the KSail cluster configuration"
		}

		if kindProp, ok := schema.Properties.Get("kind"); ok && kindProp != nil {
			kindProp.Description = "Kind of the resource (always 'Cluster' for KSail configurations)"
		}

		if specProp, ok := schema.Properties.Get("spec"); ok && specProp != nil {
			specProp.Description = "Specification of the desired cluster state"
			addSpecPropertyDescriptions(specProp)
		}
	}
}

// addSpecPropertyDescriptions adds descriptions to spec properties.
//
//nolint:cyclop // Function complexity is acceptable for property description mapping
func addSpecPropertyDescriptions(specSchema *jsonschema.Schema) {
	if specSchema.Properties == nil {
		return
	}

	descriptions := map[string]string{
		"distribution":       "Kubernetes distribution to use (Kind, K3d, or EKS)",
		"distributionConfig": "Configuration file for the distribution",
		"sourceDirectory":    "Directory containing workloads to deploy",
		"connection":         "Connection settings for the Kubernetes cluster",
		"cni":                "Container Network Interface to use",
		"csi":                "Container Storage Interface to use",
		"ingressController":  "Ingress controller to install",
		"gatewayController":  "Gateway API controller to install",
		"gitOpsEngine":       "GitOps engine to use for deployment",
		"options":            "Distribution-specific and tool-specific options",
	}

	for fieldName, description := range descriptions {
		if prop, ok := specSchema.Properties.Get(fieldName); ok && prop != nil {
			prop.Description = description
		}
	}

	// Add descriptions for connection properties
	if connProp, ok := specSchema.Properties.Get("connection"); ok &&
		connProp != nil && connProp.Properties != nil {
		if kubeconfig, ok := connProp.Properties.Get("kubeconfig"); ok && kubeconfig != nil {
			kubeconfig.Description = "Path to the kubeconfig file"
		}

		if context, ok := connProp.Properties.Get("context"); ok && context != nil {
			context.Description = "Kubernetes context of the cluster"
		}

		if timeout, ok := connProp.Properties.Get("timeout"); ok && timeout != nil {
			timeout.Description = "Timeout for cluster operations"
		}
	}
}

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
		DoNotReference:             true,
		RequiredFromJSONSchemaTags: true,
	}
	schema := reflector.Reflect(&v1alpha1.Cluster{})

	// Set the schema ID and title
	schema.ID = "https://raw.githubusercontent.com/devantler/ksail-go/main/schemas/ksail-cluster-schema.json"
	schema.Title = "KSail Cluster Configuration"
	schema.Description = "Schema for KSail cluster configuration (ksail.yaml)"

	// Mark apiVersion and kind as required fields
	// These are metadata fields that are validated by the KSail validator
	schema.Required = []string{"apiVersion", "kind"}

	// Add descriptions to properties
	addPropertyDescriptions(schema)

	// Marshal to JSON with indentation
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling schema: %w", err)
	}

	// Determine output path - default to schemas/ksail-cluster-schema.json
	outputPath := "schemas/ksail-cluster-schema.json"
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
