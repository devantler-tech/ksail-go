package config

import (
	"reflect"
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
		
		// Check if the field implements pflag.Value interface for custom types
		if pflagValue, isPflagValue := fieldPtr.(pflag.Value); isPflagValue {
			// Set default value from Viper defaults BEFORE registering the flag
			defaultVal := manager.viper.GetString(fieldPath)
			if defaultVal != "" {
				_ = pflagValue.Set(defaultVal)
			}
			
			// Use Var/VarP for custom pflag types (provides automatic validation and help)
			if shortName != "" {
				cmd.Flags().VarP(pflagValue, flagName, shortName, description)
			} else {
				cmd.Flags().Var(pflagValue, flagName, description)
			}
		} else {
			// Auto-detect type and use appropriate pflag method
			switch fieldPtr.(type) {
			case *string:
				defaultVal := manager.viper.GetString(fieldPath)
				if shortName != "" {
					cmd.Flags().StringP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().String(flagName, defaultVal, description)
				}
			case *bool:
				defaultVal := manager.viper.GetBool(fieldPath)
				if shortName != "" {
					cmd.Flags().BoolP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Bool(flagName, defaultVal, description)
				}
			case *int:
				defaultVal := manager.viper.GetInt(fieldPath)
				if shortName != "" {
					cmd.Flags().IntP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Int(flagName, defaultVal, description)
				}
			case *int32:
				defaultVal := manager.viper.GetInt32(fieldPath)
				if shortName != "" {
					cmd.Flags().Int32P(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Int32(flagName, defaultVal, description)
				}
			case *int64:
				defaultVal := manager.viper.GetInt64(fieldPath)
				if shortName != "" {
					cmd.Flags().Int64P(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Int64(flagName, defaultVal, description)
				}
			case *uint:
				defaultVal := manager.viper.GetUint(fieldPath)
				if shortName != "" {
					cmd.Flags().UintP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Uint(flagName, defaultVal, description)
				}
			case *uint32:
				defaultVal := manager.viper.GetUint32(fieldPath)
				if shortName != "" {
					cmd.Flags().Uint32P(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Uint32(flagName, defaultVal, description)
				}
			case *uint64:
				defaultVal := manager.viper.GetUint64(fieldPath)
				if shortName != "" {
					cmd.Flags().Uint64P(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Uint64(flagName, defaultVal, description)
				}
			case *float32:
				defaultVal := manager.viper.GetFloat64(fieldPath) // Viper only has Float64
				if shortName != "" {
					cmd.Flags().Float32P(flagName, shortName, float32(defaultVal), description)
				} else {
					cmd.Flags().Float32(flagName, float32(defaultVal), description)
				}
			case *float64:
				defaultVal := manager.viper.GetFloat64(fieldPath)
				if shortName != "" {
					cmd.Flags().Float64P(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Float64(flagName, defaultVal, description)
				}
			case *time.Duration:
				defaultVal := manager.viper.GetDuration(fieldPath)
				if shortName != "" {
					cmd.Flags().DurationP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Duration(flagName, defaultVal, description)
				}
			case *[]string:
				defaultVal := manager.viper.GetStringSlice(fieldPath)
				if shortName != "" {
					cmd.Flags().StringSliceP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().StringSlice(flagName, defaultVal, description)
				}
			case *[]int:
				defaultVal := manager.viper.GetIntSlice(fieldPath)
				if shortName != "" {
					cmd.Flags().IntSliceP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().IntSlice(flagName, defaultVal, description)
				}
			case *metav1.Duration:
				// Handle metav1.Duration specially as it's not a standard Duration
				defaultVal := manager.viper.GetDuration(fieldPath)
				if shortName != "" {
					cmd.Flags().DurationP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().Duration(flagName, defaultVal, description)
				}
			default:
				// Fallback to string for unknown types
				defaultVal := manager.viper.GetString(fieldPath)
				if shortName != "" {
					cmd.Flags().StringP(flagName, shortName, defaultVal, description)
				} else {
					cmd.Flags().String(flagName, defaultVal, description)
				}
			}
		}

		// Bind flag to the hierarchical path (for consistent config file access)
		// This ensures CLI flags, environment variables, and config files all use the same key
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