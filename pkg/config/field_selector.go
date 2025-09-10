package config

import (
	"reflect"
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FieldSelector represents a type-safe field selector for auto-binding.
// It provides compile-time safety by referencing actual struct fields.
type FieldSelector[T any] struct {
	selector     func(*T) any
	description  string
	defaultValue any
}

// AddFlagFromField returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
// Optionally accepts a description as the second parameter and a default value as the third parameter.
//
// Usage:
//
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution })
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, "Custom description")
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, "Custom description", v1alpha1.DistributionKind)
func AddFlagFromField[T any](selector func(*T) any, descAndDefault ...any) FieldSelector[T] {
	desc := ""
	var defaultValue any

	for i, arg := range descAndDefault {
		switch i {
		case 0:
			if s, ok := arg.(string); ok {
				desc = s
			}
		case 1:
			defaultValue = arg
		}
	}

	return FieldSelector[T]{
		selector:     selector,
		description:  desc,
		defaultValue: defaultValue,
	}
}

// AddFlagsFromFields creates field selectors from a function that provides field references.
// This provides compile-time safety with zero maintenance overhead and no global variables.
// Supports multiple modes:
//
//  1. Without descriptions - each item in the array is a field reference:
//     config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//     return []any{&c.Spec.Distribution, &c.Spec.SourceDirectory}
//     })
//
//  2. With descriptions - each field reference is followed by its description string:
//     config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//     return []any{
//     &c.Spec.Distribution, "Kubernetes distribution to use (EKS, K3d, Kind [default], Tind)",
//     &c.Spec.SourceDirectory, "Directory containing workloads to deploy",
//     }
//     })
//
//  3. With descriptions and defaults - field, description, default value:
//     config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//     return []any{
//     &c.Spec.Distribution, "Kubernetes distribution to use", v1alpha1.DistributionKind,
//     &c.Spec.SourceDirectory, "Directory containing workloads", "k8s",
//     }
//     })
func AddFlagsFromFields(
	fieldSelector func(*v1alpha1.Cluster) []any,
) []FieldSelector[v1alpha1.Cluster] {
	// Create a reference instance for field discovery
	ref := &v1alpha1.Cluster{}
	items := fieldSelector(ref)

	var selectors []FieldSelector[v1alpha1.Cluster]

	// Detect the pattern based on types in the array
	i := 0
	for i < len(items) {
		fieldPtr := items[i]
		
		// Start with field only
		selector := FieldSelector[v1alpha1.Cluster]{
			selector:     createFieldSelectorFromPointer(fieldPtr, ref),
			description:  "",
			defaultValue: nil,
		}
		
		// Check if next item is a description (string)
		if i+1 < len(items) {
			if desc, ok := items[i+1].(string); ok {
				selector.description = desc
				i++ // consume description
				
				// Check if there's a default value after description
				if i+1 < len(items) {
					// If next item is not a pointer (field), it's likely a default value
					nextItem := items[i+1]
					if reflect.ValueOf(nextItem).Kind() != reflect.Ptr {
						selector.defaultValue = nextItem
						i++ // consume default value
					}
				}
			}
		}
		
		selectors = append(selectors, selector)
		i++ // move to next field
	}

	return selectors
}

// createFieldSelectorFromPointer creates a field selector from a direct field pointer.
func createFieldSelectorFromPointer(
	fieldPtr any,
	ref *v1alpha1.Cluster,
) func(*v1alpha1.Cluster) any {
	return func(c *v1alpha1.Cluster) any {
		// Use reflection to find the field path and return the corresponding field in the target cluster
		fieldPath := getFieldPathFromPointer(fieldPtr, ref)
		if fieldPath == "" {
			return nil
		}

		return getFieldByPath(c, fieldPath)
	}
}

// getFieldPathFromPointer determines the field path from a pointer to a field in the reference cluster.
func getFieldPathFromPointer(fieldPtr any, ref *v1alpha1.Cluster) string {
	// Get the value and type of the pointer
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return ""
	}

	fieldType := fieldVal.Type().Elem() // Get the type of what the pointer points to

	// Get the address of ref and find the field path
	refVal := reflect.ValueOf(ref).Elem()
	
	return findFieldPathByType(refVal, reflect.TypeOf(ref).Elem(), fieldType, "")
}

// getFieldByPath returns a pointer to the field at the specified path in the cluster.
func getFieldByPath(cluster *v1alpha1.Cluster, path string) any {
	// Split the path into components
	parts := strings.Split(path, ".")

	// Start with the cluster value
	current := reflect.ValueOf(cluster).Elem()

	// Navigate to the target field
	for _, part := range parts {
		// Find the field by name (case-insensitive)
		fieldName := ""

		for i := 0; i < current.NumField(); i++ {
			field := current.Type().Field(i)
			if strings.EqualFold(field.Name, part) {
				fieldName = field.Name

				break
			}
		}

		if fieldName == "" {
			return nil
		}

		current = current.FieldByName(fieldName)
		if !current.IsValid() {
			return nil
		}
	}

	// Return a pointer to the field
	if current.CanAddr() {
		return current.Addr().Interface()
	}

	return nil
}

// ConfigDefault represents a configuration default with a type-safe field selector and default value.
type ConfigDefault struct {
	Path         string
	SetValue     func(*v1alpha1.Cluster, any)
	DefaultValue any
}

// GetConfigDefaults returns all configuration defaults using type-safe field setters.
// This replaces the hardcoded path mappings and provides type-safe default configuration.
func GetConfigDefaults() []ConfigDefault {
	return []ConfigDefault{
		// Metadata defaults
		{
			Path:         "metadata.name",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Metadata.Name = val.(string) },
			DefaultValue: "ksail-default",
		},

		// Spec defaults
		{
			Path:         "spec.distributionconfig",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.DistributionConfig = val.(string) },
			DefaultValue: "kind.yaml",
		},
		{
			Path:         "spec.sourcedirectory",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.SourceDirectory = val.(string) },
			DefaultValue: "k8s",
		},
		{
			Path:         "spec.distribution",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.Distribution = val.(v1alpha1.Distribution) },
			DefaultValue: v1alpha1.DistributionKind,
		},
		{
			Path:         "spec.reconciliationtool",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.ReconciliationTool = val.(v1alpha1.ReconciliationTool) },
			DefaultValue: v1alpha1.ReconciliationToolKubectl,
		},
		{
			Path:         "spec.cni",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.CNI = val.(v1alpha1.CNI) },
			DefaultValue: v1alpha1.CNIDefault,
		},
		{
			Path:         "spec.csi",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.CSI = val.(v1alpha1.CSI) },
			DefaultValue: v1alpha1.CSIDefault,
		},
		{
			Path:         "spec.ingresscontroller",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.IngressController = val.(v1alpha1.IngressController) },
			DefaultValue: v1alpha1.IngressControllerDefault,
		},
		{
			Path:         "spec.gatewaycontroller",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.GatewayController = val.(v1alpha1.GatewayController) },
			DefaultValue: v1alpha1.GatewayControllerDefault,
		},

		// Connection defaults
		{
			Path:         "spec.connection.kubeconfig",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.Connection.Kubeconfig = val.(string) },
			DefaultValue: "~/.kube/config",
		},
		{
			Path:         "spec.connection.context",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.Connection.Context = val.(string) },
			DefaultValue: "kind-ksail-default",
		},
		{
			Path:         "spec.connection.timeout",
			SetValue:     func(c *v1alpha1.Cluster, val any) { c.Spec.Connection.Timeout = val.(metav1.Duration) },
			DefaultValue: metav1.Duration{Duration: 5 * time.Minute},
		},
	}
}
