package config

import (
	"reflect"
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// bindAllFields automatically discovers and binds all fields from v1alpha1.Cluster as CLI flags.
func bindAllFields(cmd *cobra.Command, manager *Manager) {
	// Create a dummy cluster to introspect all available fields
	dummy := &v1alpha1.Cluster{}

	// Use reflection to discover all bindable fields
	allFields := discoverAllFields(dummy, reflect.ValueOf(dummy).Elem(), reflect.TypeOf(*dummy), "")

	for _, fieldInfo := range allFields {
		// Convert hierarchical path to kebab-case CLI flag
		flagName := pathToFlagName(fieldInfo.Path)

		// Generate description
		description := generateFieldDescription(fieldInfo.Path)

		// Add shortname flag if appropriate
		shortName := generateShortName(flagName)
		if shortName != "" {
			cmd.Flags().StringP(flagName, shortName, "", description)
		} else {
			// Add string flag without shortname
			cmd.Flags().String(flagName, "", description)
		}

		// Bind to both the hierarchical path (for config files) and the flat flag name (for CLI/env)
		_ = manager.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
		_ = manager.viper.BindPFlag(fieldInfo.Path, cmd.Flags().Lookup(flagName))
	}
}


// fieldInfo represents information about a discoverable field.
type fieldInfo struct {
	Path string
	Type reflect.Type
}

// discoverAllFields recursively discovers all bindable fields in a struct.
func discoverAllFields(
	rootStruct any,
	structVal reflect.Value,
	structType reflect.Type,
	prefix string,
) []fieldInfo {
	var fields []fieldInfo

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Skip embedded types (like TypeMeta) that we don't want as CLI flags
		if fieldType.Anonymous {
			continue
		}

		// Build the current field path
		var currentPath string
		if prefix == "" {
			currentPath = fieldType.Name
		} else {
			currentPath = prefix + "." + fieldType.Name
		}

		// If this is a struct (but not a special type like time.Duration), recurse into it
		if field.Kind() == reflect.Struct && !isSpecialType(field.Type()) {
			nestedFields := discoverAllFields(rootStruct, field, field.Type(), currentPath)
			fields = append(fields, nestedFields...)
		} else {
			// This is a bindable field
			fields = append(fields, fieldInfo{
				Path: currentPath,
				Type: field.Type(),
			})
		}
	}

	return fields
}

// isSpecialType checks if a type should be treated as a primitive rather than recursed into.
func isSpecialType(t reflect.Type) bool {
	// Special types that should not be recursed into
	specialTypes := []string{
		"time.Duration",
		"metav1.Duration",
		"metav1.Time",
		"metav1.ObjectMeta",
	}

	fullTypeName := t.PkgPath() + "." + t.Name()
	for _, special := range specialTypes {
		if strings.Contains(fullTypeName, special) || t.Name() == special {
			return true
		}
	}

	return false
}

// bindFieldSelectors automatically discovers and binds CLI flags for the specified field selectors.
func bindFieldSelectors(
	cmd *cobra.Command,
	manager *Manager,
	fieldSelectors []FieldSelector[v1alpha1.Cluster],
) {
	// Create a dummy cluster to introspect field paths
	dummy := &v1alpha1.Cluster{}

	for _, fieldSelector := range fieldSelectors {
		// Get the field reference from the selector
		fieldPtr := fieldSelector.selector(dummy)

		// Handle special CLI-only flags that return nil
		if fieldPtr == nil {
			// For now, we'll skip special fields - commands can add them manually if needed
			// This eliminates the need for hardcoded special field detection
			continue
		}

		// Use reflection to discover the field path
		fieldPath := getFieldPath(dummy, fieldPtr)
		if fieldPath == "" {
			continue
		}

		// Convert hierarchical path to kebab-case CLI flag
		flagName := pathToFlagName(fieldPath)

		// Use embedded description if provided, otherwise generate default
		description := fieldSelector.description
		if description == "" {
			description = generateFieldDescription(fieldPath)
		}

		// Add shortname flag if appropriate
		shortName := generateShortName(flagName)
		if shortName != "" {
			cmd.Flags().StringP(flagName, shortName, "", description)
		} else {
			// Add string flag without shortname
			cmd.Flags().String(flagName, "", description)
		}

		// Bind to both the hierarchical path (for config files) and the flat flag name (for CLI/env)
		_ = manager.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
		_ = manager.viper.BindPFlag(fieldPath, cmd.Flags().Lookup(flagName))
	}
}



// getFieldPath uses reflection to determine the path of a field within the cluster structure.
func getFieldPath(cluster *v1alpha1.Cluster, fieldPtr any) string {
	// Get the value and type of the cluster
	clusterVal := reflect.ValueOf(cluster).Elem()
	clusterType := clusterVal.Type()

	// Convert the field pointer to a reflect.Value
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return ""
	}

	fieldAddr := fieldVal.Pointer()

	// Recursively find the field path
	return findFieldPath(clusterVal, clusterType, fieldAddr, "")
}

// findFieldPath recursively searches for a field's path in a struct.
func findFieldPath(
	structVal reflect.Value,
	structType reflect.Type,
	targetAddr uintptr,
	prefix string,
) string {
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanAddr() {
			continue
		}

		// Build the current field path
		var currentPath string
		if prefix == "" {
			currentPath = fieldType.Name
		} else {
			currentPath = prefix + "." + fieldType.Name
		}

		// Check if this field's address matches our target
		if field.CanAddr() && field.Addr().Pointer() == targetAddr {
			return currentPath
		}

		// If this is a struct, recurse into it
		if field.Kind() == reflect.Struct && !isTimeType(field.Type()) {
			if result := findFieldPath(field, field.Type(), targetAddr, currentPath); result != "" {
				return result
			}
		}
	}

	return ""
}

// isTimeType checks if a type is a time-related type that shouldn't be recursed into.
func isTimeType(t reflect.Type) bool {
	return t == reflect.TypeOf(time.Time{}) || t == reflect.TypeOf(metav1.Duration{})
}

// pathToFlagName converts a hierarchical field path to a CLI flag name using the last field only.
// E.g., "metadata.name" -> "name", "spec.connection.kubeconfig" -> "kubeconfig", "spec.csi" -> "csi"
// Uppercase fields are converted to lowercase with proper kebab-case conversion.
// E.g., "spec.CSI" -> "csi", "spec.connection.IPConfig" -> "ip-config".
func pathToFlagName(path string) string {
	// Get the last part of the path
	parts := strings.Split(path, ".")
	lastPart := parts[len(parts)-1]

	// Convert camelCase and PascalCase to kebab-case
	return camelToKebab(lastPart)
}

// camelToKebab converts camelCase/PascalCase strings to kebab-case.
// E.g., "IPConfig" -> "ip-config", "CSI" -> "csi", "sourceDirectory" -> "source-directory".
func camelToKebab(s string) string {
	var result strings.Builder

	for i, r := range s {
		if i > 0 && isUpper(r) &&
			(i == len(s)-1 || !isUpper(rune(s[i+1])) || (i > 0 && !isUpper(rune(s[i-1])))) {
			result.WriteByte('-')
		}

		result.WriteRune(toLower(r))
	}

	return result.String()
}

// generateShortName generates a short flag name based on the naming rules.
// Only creates shortnames for flags longer than 3 characters.
// Uses first letter for simple names, or first letters of each word for kebab-case names.
// E.g., "csi" -> no shortname, "distribution" -> "d", "source-directory" -> "s".
// Note: Cobra only supports single-character shortnames, so we use the first letter of the first word.
func generateShortName(longName string) string {
	// Don't create shortnames for flags 3 chars or shorter
	if len(longName) <= 3 {
		return ""
	}

	// For simple names or kebab-case names, use just the first letter
	if len(longName) > 0 {
		return string(longName[0])
	}

	return ""
}

// isUpper checks if a rune is uppercase.
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// toLower converts a rune to lowercase.
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}

	return r
}

// generateFieldDescription generates a human-readable description for a configuration field.
func generateFieldDescription(fieldPath string) string {
	// Generate a default description based on the field path
	parts := strings.Split(fieldPath, ".")
	lastPart := parts[len(parts)-1]

	return "Configure " + strings.ReplaceAll(lastPart, "_", " ")
}


