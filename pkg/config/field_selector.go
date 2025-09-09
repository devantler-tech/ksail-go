package config

import (
	"reflect"
	"strings"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

// FieldSelector represents a type-safe field selector for auto-binding.
// It provides compile-time safety by referencing actual struct fields.
type FieldSelector[T any] func(*T) any

// Field returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
func Field[T any](selector func(*T) any) FieldSelector[T] {
	return selector
}

// Ref provides a v1alpha1.Cluster instance for direct field references.
// This eliminates boilerplate while maintaining compile-time safety.
//
// Usage:
//
//	config.Fields(&config.Ref.Spec.Distribution, &config.Ref.Spec.SourceDirectory)
var Ref = &v1alpha1.Cluster{}

// Fields creates field selectors from direct field pointers.
// This provides compile-time safety with zero maintenance overhead.
//
// Usage:
//
//	config.Fields(&config.Ref.Spec.Distribution, &config.Ref.Spec.SourceDirectory)
func Fields(fieldPtrs ...any) []FieldSelector[v1alpha1.Cluster] {
	var selectors []FieldSelector[v1alpha1.Cluster]

	for _, fieldPtr := range fieldPtrs {
		selector := createFieldSelectorFromPointer(fieldPtr)
		selectors = append(selectors, selector)
	}
	return selectors
}

// createFieldSelectorFromPointer creates a field selector from a direct field pointer.
func createFieldSelectorFromPointer(fieldPtr any) FieldSelector[v1alpha1.Cluster] {
	return func(c *v1alpha1.Cluster) any {
		// Use reflection to find the field path and return the corresponding field in the target cluster
		fieldPath := getFieldPathFromPointer(fieldPtr)
		if fieldPath == "" {
			return nil
		}
		return getFieldByPath(c, fieldPath)
	}
}

// getFieldPathFromPointer determines the field path from a pointer to a field in Ref.
func getFieldPathFromPointer(fieldPtr any) string {
	// Get the value and type of the pointer
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return ""
	}
	fieldAddr := fieldVal.Pointer()

	// Get the address of Ref and find the field path
	refVal := reflect.ValueOf(Ref).Elem()
	refType := refVal.Type()

	return findFieldPath(refVal, refType, fieldAddr, "")
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
			if strings.ToLower(field.Name) == strings.ToLower(part) {
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
