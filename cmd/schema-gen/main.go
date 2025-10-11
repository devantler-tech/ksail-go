// Package main implements the JSON schema generator for KSail configuration.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/invopop/jsonschema"
)

const (
	dirPermissions  = 0o750 // Directory permissions (rwxr-x---)
	filePermissions = 0o600 // File permissions (rw-------)
)

// addPropertyDescriptions adds human-readable descriptions to schema properties.
// Descriptions are dynamically extracted from field selectors in the config manager.
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
// Descriptions are extracted dynamically from field selectors defined in the config manager.
func addSpecPropertyDescriptions(specSchema *jsonschema.Schema) {
	if specSchema.Properties == nil {
		return
	}

	// Get descriptions from field selectors
	fieldDescriptions := extractFieldDescriptions()

	// Apply descriptions to properties
	for fieldPath, description := range fieldDescriptions {
		if prop, ok := specSchema.Properties.Get(fieldPath); ok && prop != nil {
			prop.Description = description
		}
	}

	// Add descriptions for nested connection properties
	addConnectionPropertyDescriptions(specSchema)
}

// extractFieldDescriptions extracts field descriptions from all available field selectors.
func extractFieldDescriptions() map[string]string {
	descriptions := make(map[string]string)

	// Create a sample cluster to use with field selectors
	cluster := &v1alpha1.Cluster{}

	// Get all standard field selectors
	selectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.DefaultDistributionFieldSelector(),
		configmanager.DefaultDistributionConfigFieldSelector(),
		configmanager.StandardSourceDirectoryFieldSelector(),
		configmanager.DefaultContextFieldSelector(),
	}

	// Extract field paths and descriptions from selectors
	for _, selector := range selectors {
		if selector.Description == "" {
			continue
		}

		// Get the field pointer from the selector
		fieldPtr := selector.Selector(cluster)

		// Determine the field path by inspecting the pointer
		fieldPath := getFieldPath(cluster, fieldPtr)
		if fieldPath != "" {
			descriptions[fieldPath] = selector.Description
		}
	}

	// Add additional descriptions for fields not covered by standard selectors
	// These are fields that don't have CLI flags but need descriptions
	additionalDescriptions := map[string]string{
		"connection":        "Connection settings for the Kubernetes cluster",
		"cni":               "Container Network Interface to use",
		"csi":               "Container Storage Interface to use",
		"ingressController": "Ingress controller to install",
		"gatewayController": "Gateway API controller to install",
		"gitOpsEngine":      "GitOps engine to use for deployment",
		"options":           "Distribution-specific and tool-specific options",
	}

	for key, desc := range additionalDescriptions {
		if _, exists := descriptions[key]; !exists {
			descriptions[key] = desc
		}
	}

	return descriptions
}

// getFieldPath determines the JSON field name for a given field pointer.
func getFieldPath(cluster *v1alpha1.Cluster, fieldPtr any) string {
	// Map field pointers to their JSON field names
	switch fieldPtr {
	case &cluster.Spec.Distribution:
		return "distribution"
	case &cluster.Spec.DistributionConfig:
		return "distributionConfig"
	case &cluster.Spec.SourceDirectory:
		return "sourceDirectory"
	case &cluster.Spec.Connection.Context:
		return "context"
	case &cluster.Spec.Connection.Kubeconfig:
		return "kubeconfig"
	case &cluster.Spec.Connection.Timeout:
		return "timeout"
	}

	// If we can't determine the field path, try reflection as a fallback
	return getFieldPathByReflection(cluster, fieldPtr)
}

// getFieldPathByReflection uses reflection to find the field path.
func getFieldPathByReflection(cluster *v1alpha1.Cluster, fieldPtr any) string {
	clusterVal := reflect.ValueOf(cluster).Elem()
	ptrVal := reflect.ValueOf(fieldPtr)

	// Check if it's a pointer
	if ptrVal.Kind() != reflect.Ptr {
		return ""
	}

	// Try to find the field in Spec
	specVal := clusterVal.FieldByName("Spec")
	if !specVal.IsValid() {
		return ""
	}

	return findFieldInStruct(specVal, ptrVal.Pointer())
}

//nolint:nestif,intrange // Reflection code requires complex nested conditions
func findFieldInStruct(structVal reflect.Value, targetPtr uintptr) string {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Check if this field's address matches
		if field.CanAddr() && field.Addr().Pointer() == targetPtr {
			// Get JSON tag name
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag != "" {
				// Parse the tag (e.g., "fieldName,omitzero")
				if idx := len(jsonTag); idx > 0 {
					for j, c := range jsonTag {
						if c == ',' {
							return jsonTag[:j]
						}
					}

					return jsonTag
				}
			}

			return fieldType.Name
		}

		// Recursively search in nested structs
		if field.Kind() == reflect.Struct {
			result := findFieldInStruct(field, targetPtr)
			if result != "" {
				return result
			}
		}
	}

	return ""
}

// addConnectionPropertyDescriptions adds descriptions for connection properties.
//
//nolint:cyclop // Function complexity is acceptable for property description mapping
func addConnectionPropertyDescriptions(specSchema *jsonschema.Schema) {
	connProp, ok := specSchema.Properties.Get("connection")
	if !ok || connProp == nil || connProp.Properties == nil {
		return
	}

	// Get the context field selector for its description
	contextSelector := configmanager.DefaultContextFieldSelector()

	if kubeconfig, ok := connProp.Properties.Get("kubeconfig"); ok && kubeconfig != nil {
		kubeconfig.Description = "Path to the kubeconfig file"
	}

	if context, ok := connProp.Properties.Get("context"); ok && context != nil {
		// Use the description from the field selector
		if contextSelector.Description != "" {
			context.Description = contextSelector.Description
		}
	}

	if timeout, ok := connProp.Properties.Get("timeout"); ok && timeout != nil {
		timeout.Description = "Timeout for cluster operations"
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

	// Add descriptions to properties dynamically from field selectors
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
