package config

import (
	"reflect"
	"strings"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

// FieldSelector represents a type-safe field selector for auto-binding.
// It provides compile-time safety by referencing actual struct fields.
type FieldSelector[T any] struct {
	selector    func(*T) any
	description string
}

// Field returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
// Optionally accepts a description as the second parameter.
//
// Usage:
//   Field(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution })
//   Field(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, "Custom description")
func Field[T any](selector func(*T) any, description ...string) FieldSelector[T] {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	return FieldSelector[T]{
		selector:    selector,
		description: desc,
	}
}

// Fields creates field selectors from a function that provides field references.
// This provides compile-time safety with zero maintenance overhead and no global variables.
// Supports two modes:
//
// 1. Without descriptions - each item in the array is a field reference:
//    config.Fields(func(c *v1alpha1.Cluster) []any {
//        return []any{&c.Spec.Distribution, &c.Spec.SourceDirectory}
//    })
//
// 2. With descriptions - each field reference is followed by its description string:
//    config.Fields(func(c *v1alpha1.Cluster) []any {
//        return []any{
//            &c.Spec.Distribution, "Kubernetes distribution to use (EKS, K3d, Kind [default], Tind)",
//            &c.Spec.SourceDirectory, "Directory containing workloads to deploy",
//        }
//    })
func Fields(fieldSelector func(*v1alpha1.Cluster) []any) []FieldSelector[v1alpha1.Cluster] {
	// Create a reference instance for field discovery
	ref := &v1alpha1.Cluster{}
	items := fieldSelector(ref)

	var selectors []FieldSelector[v1alpha1.Cluster]

	// Check if we have descriptions by looking for string items
	hasDescriptions := false
	for i := 1; i < len(items); i += 2 {
		if _, ok := items[i].(string); ok {
			hasDescriptions = true
			break
		}
	}

	if hasDescriptions {
		// Process items in pairs: field pointer, description
		for i := 0; i < len(items); i += 2 {
			if i+1 >= len(items) {
				// Odd number of items, treat as field without description
				fieldPtr := items[i]
				selector := createFieldSelectorFromPointer(fieldPtr, ref)
				selectors = append(selectors, FieldSelector[v1alpha1.Cluster]{
					selector:    selector,
					description: "",
				})
				continue
			}

			fieldPtr := items[i]
			desc, ok := items[i+1].(string)
			if !ok {
				// Not a string description, treat both as fields without descriptions
				selector1 := createFieldSelectorFromPointer(fieldPtr, ref)
				selectors = append(selectors, FieldSelector[v1alpha1.Cluster]{
					selector:    selector1,
					description: "",
				})

				selector2 := createFieldSelectorFromPointer(items[i+1], ref)
				selectors = append(selectors, FieldSelector[v1alpha1.Cluster]{
					selector:    selector2,
					description: "",
				})
				continue
			}

			// Valid field + description pair
			selector := createFieldSelectorFromPointer(fieldPtr, ref)
			selectors = append(selectors, FieldSelector[v1alpha1.Cluster]{
				selector:    selector,
				description: desc,
			})
		}
	} else {
		// Process all items as field pointers without descriptions
		for _, fieldPtr := range items {
			selector := createFieldSelectorFromPointer(fieldPtr, ref)
			selectors = append(selectors, FieldSelector[v1alpha1.Cluster]{
				selector:    selector,
				description: "",
			})
		}
	}

	return selectors
}

// createFieldSelectorFromPointer creates a field selector from a direct field pointer.
func createFieldSelectorFromPointer(fieldPtr any, ref *v1alpha1.Cluster) func(*v1alpha1.Cluster) any {
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

	fieldAddr := fieldVal.Pointer()

	// Get the address of ref and find the field path
	refVal := reflect.ValueOf(ref).Elem()
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
