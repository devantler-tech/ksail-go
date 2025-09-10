package config

import (
	"reflect"
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultDurationTimeout = 5 * time.Minute

// FieldSelector represents a type-safe field selector for auto-binding.
// It provides compile-time safety by referencing actual struct fields.
type FieldSelector[T any] struct {
	selector     func(*T) any
	description  string
	defaultValue any
}

// AddFlagFromField returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
// Requires a default value as the second parameter, optionally accepts a description as the third parameter.
//
// Usage:
//
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, v1alpha1.DistributionKind)
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//		v1alpha1.DistributionKind, "Custom description")
func AddFlagFromField[T any](
	selector func(*T) any,
	defaultValue any,
	description ...string,
) FieldSelector[T] {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return FieldSelector[T]{
		selector:     selector,
		description:  desc,
		defaultValue: defaultValue,
	}
}

// AddFlagsFromFields creates field selectors from a function that provides field references.
// This provides compile-time safety with zero maintenance overhead and no global variables.
// Each field must be followed by its default value and optionally by a description string.
//
//	Usage with defaults only:
//	   config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//	       return []any{
//	           &c.Spec.Distribution, v1alpha1.DistributionKind,
//	           &c.Spec.SourceDirectory, "k8s",
//	       }
//	   })
//
//	Usage with defaults and descriptions:
//	   config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//	       return []any{
//	           &c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to use",
//	           &c.Spec.SourceDirectory, "k8s", "Directory containing workloads to deploy",
//	       }
//	   })
func AddFlagsFromFields(
	fieldSelector func(*v1alpha1.Cluster) []any,
) []FieldSelector[v1alpha1.Cluster] {
	// Create a reference instance for field discovery
	ref := &v1alpha1.Cluster{} //nolint:exhaustruct // Only used for reflection, empty is correct
	items := fieldSelector(ref)

	var selectors []FieldSelector[v1alpha1.Cluster]

	// Each field must have at least a default value, and optionally a description
	itemIndex := 0
	for itemIndex < len(items) {
		if itemIndex+1 >= len(items) {
			break // Need at least field and default value
		}

		fieldPtr := items[itemIndex]
		defaultValue := items[itemIndex+1]
		itemIndex += 2

		// Check if next item is a description (string)
		description := ""

		if itemIndex < len(items) {
			if desc, ok := items[itemIndex].(string); ok {
				description = desc
				itemIndex++ // consume description
			}
		}

		selector := FieldSelector[v1alpha1.Cluster]{
			selector:     createFieldSelectorFromPointer(fieldPtr, ref),
			description:  description,
			defaultValue: defaultValue,
		}

		selectors = append(selectors, selector)
	}

	return selectors
}

// createFieldSelectorFromPointer creates a field selector from a direct field pointer.
func createFieldSelectorFromPointer(
	fieldPtr any,
	ref *v1alpha1.Cluster,
) func(*v1alpha1.Cluster) any {
	return func(cluster *v1alpha1.Cluster) any {
		// Use reflection to find the field path and return the corresponding field in the target cluster
		fieldPath := getFieldPathFromPointer(fieldPtr, ref)
		if fieldPath == "" {
			return nil
		}

		return getFieldByPath(cluster, fieldPath)
	}
}

// getFieldPathFromPointer determines the field path from a pointer to a field in the reference cluster.
func getFieldPathFromPointer(fieldPtr any, ref *v1alpha1.Cluster) string {
	// Get reflection info about the field pointer
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return ""
	}

	// Get the address and type of the field
	fieldAddr := fieldVal.Pointer()
	fieldType := fieldVal.Type()

	// Walk the cluster structure to find the field with this address and type
	refVal := reflect.ValueOf(ref).Elem()

	// Get path in original case then convert to lowercase for Viper
	originalPath := findFieldPathByAddressAndType(
		refVal,
		reflect.TypeOf(ref).Elem(),
		fieldAddr,
		fieldType,
		"",
		false,
	)

	return strings.ToLower(originalPath)
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

		for fieldIndex := range current.NumField() {
			field := current.Type().Field(fieldIndex)
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

// convertValueToFieldType converts a value from Viper to the appropriate field type.
func convertValueToFieldType(value any, targetType reflect.Type) any {
	if value == nil {
		return nil
	}

	// Handle metav1.Duration specially - it has a time.Duration field
	if result := handleMetav1Duration(value, targetType); result != nil {
		return result
	}

	// Handle string values from Viper - try pflag.Value interface first for more elegant enum handling
	if strVal, ok := value.(string); ok {
		if result := handleStringConversion(strVal, targetType); result != nil {
			return result
		}
	}

	// Handle other types (direct assignment)
	return handleDirectConversion(value, targetType)
}

// handleMetav1Duration handles conversion for metav1.Duration type.
func handleMetav1Duration(value any, targetType reflect.Type) any {
	if targetType != reflect.TypeOf(metav1.Duration{}) { //nolint:exhaustruct
		return nil
	}

	switch val := value.(type) {
	case time.Duration:
		return metav1.Duration{Duration: val}
	case string:
		duration, err := time.ParseDuration(val)
		if err == nil {
			return metav1.Duration{Duration: duration}
		}

		return metav1.Duration{Duration: defaultDurationTimeout}
	case metav1.Duration:
		return val
	}

	return metav1.Duration{Duration: defaultDurationTimeout}
}

// handleStringConversion handles string value conversions using pflag.Value interface.
func handleStringConversion(strVal string, targetType reflect.Type) any {
	// Check if target type implements pflag.Value interface
	targetValuePtr := reflect.New(targetType)
	if pflagValue, implements := targetValuePtr.Interface().(pflag.Value); implements {
		// Use pflag.Value interface for elegant enum conversion
		err := pflagValue.Set(strVal)
		if err == nil {
			return targetValuePtr.Elem().Interface()
		}
		// Fallback to type-specific defaults on error
		return getEnumDefault(targetType)
	}

	// Fallback to direct type conversion for non-pflag.Value types
	if targetType == reflect.TypeOf("") {
		return strVal
	}

	return nil
}

// getEnumDefault returns default values for enum types.
func getEnumDefault(targetType reflect.Type) any {
	switch targetType {
	case reflect.TypeOf(v1alpha1.Distribution("")):
		return v1alpha1.DistributionKind
	case reflect.TypeOf(v1alpha1.ReconciliationTool("")):
		return v1alpha1.ReconciliationToolKubectl
	case reflect.TypeOf(v1alpha1.CNI("")):
		return v1alpha1.CNIDefault
	case reflect.TypeOf(v1alpha1.CSI("")):
		return v1alpha1.CSIDefault
	case reflect.TypeOf(v1alpha1.IngressController("")):
		return v1alpha1.IngressControllerDefault
	case reflect.TypeOf(v1alpha1.GatewayController("")):
		return v1alpha1.GatewayControllerDefault
	}

	return nil
}

// handleDirectConversion handles direct type conversions.
func handleDirectConversion(value any, targetType reflect.Type) any {
	// Handle other types (direct assignment)
	if reflect.TypeOf(value) == targetType {
		return value
	}

	// Fallback: try to convert using reflection
	valueVal := reflect.ValueOf(value)
	if valueVal.Type().ConvertibleTo(targetType) {
		return valueVal.Convert(targetType).Interface()
	}

	return value
}
