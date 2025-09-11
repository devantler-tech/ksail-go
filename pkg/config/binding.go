// Package config provides centralized configuration management using Viper.
// This file contains field selector binding functionality for automatic CLI flag creation.
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
	dummy := &v1alpha1.Cluster{} //nolint:exhaustruct // Only used for reflection, empty is correct
	usedShorthands := make(map[string]bool)

	for _, fieldSelector := range fieldSelectors {
		bindSingleFieldSelector(cmd, manager, dummy, fieldSelector, usedShorthands)
	}
}

// bindSingleFieldSelector binds a single field selector to the command.
func bindSingleFieldSelector(
	cmd *cobra.Command,
	manager *Manager,
	dummy *v1alpha1.Cluster,
	fieldSelector FieldSelector[v1alpha1.Cluster],
	usedShorthands map[string]bool,
) {
	// Get the field reference from the selector
	fieldPtr := fieldSelector.selector(dummy)

	// Handle special CLI-only flags that return nil
	if fieldPtr == nil {
		// For now, we'll skip special fields - commands can add them manually if needed
		// This eliminates the need for hardcoded special field detection
		return
	}

	// Use reflection to discover the field path
	fieldPath := getFieldPath(dummy, fieldPtr)
	if fieldPath == "" {
		return
	}

	// Get field path with preserved case for flag name generation
	fieldPathWithCase := getFieldPathPreservingCase(dummy, fieldPtr)

	// Convert hierarchical path to kebab-case CLI flag (use case-preserving path)
	flagName := pathToFlagName(fieldPathWithCase)

	// Use embedded description if provided, otherwise generate default
	description := fieldSelector.description
	if description == "" {
		description = generateFieldDescription(fieldPathWithCase)
	}

	// Get default value from field selector or fallback to Viper
	var defaultValue any
	if fieldSelector.defaultValue != nil {
		defaultValue = fieldSelector.defaultValue
	}

	// Add shortname flag if appropriate and not conflicting
	shortName := generateShortName(flagName)
	if shortName != "" && usedShorthands[shortName] {
		shortName = "" // Avoid conflicts by not using shorthand
	}

	if shortName != "" {
		usedShorthands[shortName] = true
	}

	// Check if the field implements pflag.Value interface for custom types
	if pflagValue, isPflagValue := fieldPtr.(pflag.Value); isPflagValue {
		bindPflagValue(
			cmd, manager, pflagValue, flagName, shortName, description, fieldPath, defaultValue,
		)
	} else {
		bindStandardType(
			cmd, manager, fieldPtr, flagName, shortName, description, fieldPath, defaultValue,
		)
	}

	// Bind flag to the hierarchical path (for consistent config file access)
	// This ensures CLI flags, environment variables, and config files all use the same key
	_ = manager.viper.BindPFlag(fieldPath, cmd.Flags().Lookup(flagName))
}

// bindPflagValue binds a pflag.Value type to the command.
func bindPflagValue(
	cmd *cobra.Command,
	manager *Manager,
	pflagValue pflag.Value,
	flagName, shortName, description, fieldPath string,
	defaultValue any,
) {
	// Set default value from field selector or Viper defaults
	if defaultValue != nil {
		setPflagValueDefault(pflagValue, defaultValue)
	} else {
		// Fallback to Viper defaults
		defaultVal := manager.viper.GetString(fieldPath)
		if defaultVal != "" {
			_ = pflagValue.Set(defaultVal)
		}
	}

	// Use Var/VarP for custom pflag types (provides automatic validation and help)
	if shortName != "" {
		cmd.Flags().VarP(pflagValue, flagName, shortName, description)
	} else {
		cmd.Flags().Var(pflagValue, flagName, description)
	}
}

// setPflagValueDefault sets the default value for a pflag.Value.
func setPflagValueDefault(pflagValue pflag.Value, defaultValue any) {
	// Convert custom types to string for pflag.Value.Set()
	switch val := defaultValue.(type) {
	case v1alpha1.Distribution:
		_ = pflagValue.Set(string(val))
	case v1alpha1.ReconciliationTool:
		_ = pflagValue.Set(string(val))
	case v1alpha1.CNI:
		_ = pflagValue.Set(string(val))
	case v1alpha1.CSI:
		_ = pflagValue.Set(string(val))
	case v1alpha1.IngressController:
		_ = pflagValue.Set(string(val))
	case v1alpha1.GatewayController:
		_ = pflagValue.Set(string(val))
	default:
		if str, ok := val.(string); ok {
			_ = pflagValue.Set(str)
		}
	}
}

// bindStandardType binds a standard type to the command using type detection.
func bindStandardType(
	cmd *cobra.Command,
	manager *Manager,
	fieldPtr any,
	flagName, shortName, description, fieldPath string,
	defaultValue any,
) {
	// Auto-detect type and use appropriate pflag method
	switch fieldPtr.(type) {
	case *string:
		bindStringFlag(cmd, manager, flagName, shortName, description, fieldPath, defaultValue)
	case *bool:
		bindBoolFlag(cmd, manager, flagName, shortName, description, fieldPath, defaultValue)
	case *int, *int32, *int64:
		bindIntegerFlag(cmd, manager, fieldPtr, flagName, shortName, description, fieldPath)
	case *uint, *uint32, *uint64:
		bindUnsignedIntegerFlag(cmd, manager, fieldPtr, flagName, shortName, description, fieldPath)
	case *float32, *float64:
		bindFloatFlag(cmd, manager, fieldPtr, flagName, shortName, description, fieldPath)
	case *time.Duration, *metav1.Duration:
		bindDurationTypes(cmd, manager, fieldPtr, flagName, shortName, description, fieldPath, defaultValue)
	case *[]string:
		bindStringSliceFlag(cmd, manager, flagName, shortName, description, fieldPath)
	case *[]int:
		bindIntSliceFlag(cmd, manager, flagName, shortName, description, fieldPath)
	default:
		// Fallback to string for unknown types
		bindStringFlag(cmd, manager, flagName, shortName, description, fieldPath, defaultValue)
	}
}

// bindDurationTypes binds duration types to the command.
func bindDurationTypes(
	cmd *cobra.Command,
	manager *Manager,
	fieldPtr any,
	flagName, shortName, description, fieldPath string,
	defaultValue any,
) {
	switch fieldPtr.(type) {
	case *time.Duration:
		bindDurationFlag(cmd, manager, flagName, shortName, description, fieldPath)
	case *metav1.Duration:
		bindMetav1DurationFlag(cmd, manager, flagName, shortName, description, fieldPath, defaultValue)
	}
}

// bindIntegerFlag binds integer types to the command.
func bindIntegerFlag(
	cmd *cobra.Command,
	manager *Manager,
	fieldPtr any,
	flagName, shortName, description, fieldPath string,
) {
	switch fieldPtr.(type) {
	case *int:
		bindIntFlag(cmd, manager, flagName, shortName, description, fieldPath)
	case *int32:
		bindInt32Flag(cmd, manager, flagName, shortName, description, fieldPath)
	case *int64:
		bindInt64Flag(cmd, manager, flagName, shortName, description, fieldPath)
	}
}

// bindUnsignedIntegerFlag binds unsigned integer types to the command.
func bindUnsignedIntegerFlag(
	cmd *cobra.Command,
	manager *Manager,
	fieldPtr any,
	flagName, shortName, description, fieldPath string,
) {
	switch fieldPtr.(type) {
	case *uint:
		bindUintFlag(cmd, manager, flagName, shortName, description, fieldPath)
	case *uint32:
		bindUint32Flag(cmd, manager, flagName, shortName, description, fieldPath)
	case *uint64:
		bindUint64Flag(cmd, manager, flagName, shortName, description, fieldPath)
	}
}

// bindFloatFlag binds float types to the command.
func bindFloatFlag(
	cmd *cobra.Command,
	manager *Manager,
	fieldPtr any,
	flagName, shortName, description, fieldPath string,
) {
	switch fieldPtr.(type) {
	case *float32:
		bindFloat32Flag(cmd, manager, flagName, shortName, description, fieldPath)
	case *float64:
		bindFloat64Flag(cmd, manager, flagName, shortName, description, fieldPath)
	}
}

// bindStringFlag binds a string flag to the command.
func bindStringFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
	defaultValue any,
) {
	var defaultVal string
	if defaultValue != nil {
		if str, ok := defaultValue.(string); ok {
			defaultVal = str
		}
	} else {
		defaultVal = manager.viper.GetString(fieldPath)
	}

	if shortName != "" {
		cmd.Flags().StringP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().String(flagName, defaultVal, description)
	}
}

// bindBoolFlag binds a bool flag to the command.
func bindBoolFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
	defaultValue any,
) {
	var defaultVal bool
	if defaultValue != nil {
		if b, ok := defaultValue.(bool); ok {
			defaultVal = b
		}
	} else {
		defaultVal = manager.viper.GetBool(fieldPath)
	}

	if shortName != "" {
		cmd.Flags().BoolP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Bool(flagName, defaultVal, description)
	}
}

// bindIntFlag binds an int flag to the command.
func bindIntFlag(
	cmd *cobra.Command, manager *Manager, flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetInt(fieldPath)
	if shortName != "" {
		cmd.Flags().IntP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Int(flagName, defaultVal, description)
	}
}

// bindInt32Flag binds an int32 flag to the command.
func bindInt32Flag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetInt32(fieldPath)
	if shortName != "" {
		cmd.Flags().Int32P(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Int32(flagName, defaultVal, description)
	}
}

// bindInt64Flag binds an int64 flag to the command.
func bindInt64Flag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetInt64(fieldPath)
	if shortName != "" {
		cmd.Flags().Int64P(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Int64(flagName, defaultVal, description)
	}
}

// bindUintFlag binds a uint flag to the command.
func bindUintFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetUint(fieldPath)
	if shortName != "" {
		cmd.Flags().UintP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Uint(flagName, defaultVal, description)
	}
}

// bindUint32Flag binds a uint32 flag to the command.
func bindUint32Flag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetUint32(fieldPath)
	if shortName != "" {
		cmd.Flags().Uint32P(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Uint32(flagName, defaultVal, description)
	}
}

// bindUint64Flag binds a uint64 flag to the command.
func bindUint64Flag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetUint64(fieldPath)
	if shortName != "" {
		cmd.Flags().Uint64P(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Uint64(flagName, defaultVal, description)
	}
}

// bindFloat32Flag binds a float32 flag to the command.
func bindFloat32Flag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetFloat64(fieldPath) // Viper only has Float64
	if shortName != "" {
		cmd.Flags().Float32P(flagName, shortName, float32(defaultVal), description)
	} else {
		cmd.Flags().Float32(flagName, float32(defaultVal), description)
	}
}

// bindFloat64Flag binds a float64 flag to the command.
func bindFloat64Flag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetFloat64(fieldPath)
	if shortName != "" {
		cmd.Flags().Float64P(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Float64(flagName, defaultVal, description)
	}
}

// bindDurationFlag binds a time.Duration flag to the command.
func bindDurationFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetDuration(fieldPath)
	if shortName != "" {
		cmd.Flags().DurationP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Duration(flagName, defaultVal, description)
	}
}

// bindStringSliceFlag binds a []string flag to the command.
func bindStringSliceFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetStringSlice(fieldPath)
	if shortName != "" {
		cmd.Flags().StringSliceP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().StringSlice(flagName, defaultVal, description)
	}
}

// bindIntSliceFlag binds a []int flag to the command.
func bindIntSliceFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
) {
	defaultVal := manager.viper.GetIntSlice(fieldPath)
	if shortName != "" {
		cmd.Flags().IntSliceP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().IntSlice(flagName, defaultVal, description)
	}
}

// bindMetav1DurationFlag binds a metav1.Duration flag to the command.
func bindMetav1DurationFlag(
	cmd *cobra.Command,
	manager *Manager,
	flagName, shortName, description, fieldPath string,
	defaultValue any,
) {
	// Handle metav1.Duration specially as it's not a standard Duration
	var defaultVal time.Duration
	if defaultValue != nil {
		if metaDur, ok := defaultValue.(metav1.Duration); ok {
			defaultVal = metaDur.Duration
		}
	} else {
		defaultVal = manager.viper.GetDuration(fieldPath)
	}

	if shortName != "" {
		cmd.Flags().DurationP(flagName, shortName, defaultVal, description)
	} else {
		cmd.Flags().Duration(flagName, defaultVal, description)
	}
}

// getFieldReflectionInfo extracts common reflection information from a field pointer.
func getFieldReflectionInfo(
	fieldPtr any,
) (uintptr, reflect.Type, bool) {
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return 0, nil, false
	}

	fieldAddr := fieldVal.Pointer()
	fieldType := fieldVal.Type()

	return fieldAddr, fieldType, true
}

// getFieldPath uses a combination of memory address and reflection to discover the field path.
func getFieldPath(cluster *v1alpha1.Cluster, fieldPtr any) string {
	fieldAddr, fieldType, valid := getFieldReflectionInfo(fieldPtr)
	if !valid {
		return ""
	}

	// Walk the cluster structure to find the field with this address and type
	clusterVal := reflect.ValueOf(cluster).Elem()

	// Get the field path in original case first
	originalPath := findFieldPathByAddressAndType(
		clusterVal,
		reflect.TypeOf(cluster).Elem(),
		fieldAddr,
		fieldType,
		"",
		false,
	)

	// Convert to lowercase for Viper compatibility
	return strings.ToLower(originalPath)
}

// getFieldPathPreservingCase gets the field path while preserving original case for flag name generation.
func getFieldPathPreservingCase(cluster *v1alpha1.Cluster, fieldPtr any) string {
	fieldAddr, fieldType, valid := getFieldReflectionInfo(fieldPtr)
	if !valid {
		return ""
	}

	// Walk the cluster structure to find the field with this address and type
	clusterVal := reflect.ValueOf(cluster).Elem()

	// Return the field path in original case
	return findFieldPathByAddressAndType(
		clusterVal,
		reflect.TypeOf(cluster).Elem(),
		fieldAddr,
		fieldType,
		"",
		true,
	)
}

// findFieldPathByAddressAndType recursively searches for a field's path by comparing memory addresses and types.
func findFieldPathByAddressAndType(
	structVal reflect.Value,
	structType reflect.Type,
	targetAddr uintptr,
	targetType reflect.Type,
	prefix string,
	_ bool, // preserveCase is unused
) string {
	for fieldIndex := range structVal.NumField() {
		field := structVal.Field(fieldIndex)
		fieldType := structType.Field(fieldIndex)

		// Skip unexported fields
		if !field.CanAddr() {
			continue
		}

		// Build the current field path (preserve original case for accurate discovery)
		var currentPath string
		if prefix == "" {
			currentPath = fieldType.Name
		} else {
			currentPath = prefix + "." + fieldType.Name
		}

		// Check if this field's address AND type matches our target
		if field.CanAddr() && field.Addr().Pointer() == targetAddr &&
			field.Addr().Type() == targetType {
			return currentPath
		}

		// If this is a struct, recurse into it
		if field.Kind() == reflect.Struct && !isTimeType(field.Type()) {
			result := findFieldPathByAddressAndType(
				field, field.Type(), targetAddr, targetType, currentPath, false,
			)
			if result != "" {
				return result
			}
		}
	}

	return ""
}

// isTimeType checks if a type is a time-related type that shouldn't be recursed into.
func isTimeType(t reflect.Type) bool {
	return t == reflect.TypeOf(time.Time{}) ||
		t == reflect.TypeOf(metav1.Duration{}) //nolint:exhaustruct
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

	for index, char := range s {
		if index > 0 && isUpper(char) &&
			(index == len(s)-1 || !isUpper(rune(s[index+1])) || (index > 0 && !isUpper(rune(s[index-1])))) {
			result.WriteByte('-')
		}

		result.WriteRune(toLower(char))
	}

	return result.String()
}

// generateShortName generates a short flag name based on the naming rules.
// Only creates shortnames for flags longer than 3 characters.
// Uses first letter for simple names, or first letters of each word for kebab-case names.
// E.g., "csi" -> no shortname, "distribution" -> "d", "source-directory" -> "s".
// Note: Cobra only supports single-character shortnames, so we use the first letter of the first word.
func generateShortName(longName string) string {
	const minShortnameLength = 3

	// Don't create shortnames for flags shorter than the minimum length
	if len(longName) <= minShortnameLength {
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
