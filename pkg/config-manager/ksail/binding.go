// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains field selector binding functionality for automatic CLI flag creation.
package ksail

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddFlagsFromFields adds CLI flags for all configured field selectors.
func (m *Manager) AddFlagsFromFields(cmd *cobra.Command) {
	for _, fieldSelector := range m.fieldSelectors {
		m.addFlagFromField(cmd, fieldSelector)
	}
}

// addFlagFromField adds a CLI flag for a single field selector.
func (m *Manager) addFlagFromField(
	cmd *cobra.Command,
	fieldSelector FieldSelector[v1alpha1.Cluster],
) {
	fieldPtr := fieldSelector.Selector(m.Config)
	if fieldPtr == nil {
		return
	}

	flagName := m.generateFlagName(fieldPtr)
	shorthand := m.generateShorthand(flagName)

	switch ptr := fieldPtr.(type) {
	case *v1alpha1.Distribution:
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, string(fieldSelector.DefaultValue.(v1alpha1.Distribution)), fieldSelector.Description)
	case *v1alpha1.ReconciliationTool:
		defaultTool := ""
		if fieldSelector.DefaultValue != nil {
			defaultTool = string(fieldSelector.DefaultValue.(v1alpha1.ReconciliationTool))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultTool, fieldSelector.Description)
	case *v1alpha1.CNI:
		defaultCNI := ""
		if fieldSelector.DefaultValue != nil {
			defaultCNI = string(fieldSelector.DefaultValue.(v1alpha1.CNI))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultCNI, fieldSelector.Description)
	case *v1alpha1.CSI:
		defaultCSI := ""
		if fieldSelector.DefaultValue != nil {
			defaultCSI = string(fieldSelector.DefaultValue.(v1alpha1.CSI))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultCSI, fieldSelector.Description)
	case *v1alpha1.IngressController:
		defaultIngress := ""
		if fieldSelector.DefaultValue != nil {
			defaultIngress = string(fieldSelector.DefaultValue.(v1alpha1.IngressController))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultIngress, fieldSelector.Description)
	case *v1alpha1.GatewayController:
		defaultGateway := ""
		if fieldSelector.DefaultValue != nil {
			defaultGateway = string(fieldSelector.DefaultValue.(v1alpha1.GatewayController))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultGateway, fieldSelector.Description)
	case *string:
		defaultStr := ""
		if fieldSelector.DefaultValue != nil {
			defaultStr = fieldSelector.DefaultValue.(string)
		}
		cmd.Flags().StringVarP(ptr, flagName, shorthand, defaultStr, fieldSelector.Description)
	case *bool:
		defaultBool := false
		if fieldSelector.DefaultValue != nil {
			defaultBool = fieldSelector.DefaultValue.(bool)
		}
		cmd.Flags().BoolVarP(ptr, flagName, shorthand, defaultBool, fieldSelector.Description)
	case *int:
		defaultInt := 0
		if fieldSelector.DefaultValue != nil {
			defaultInt = fieldSelector.DefaultValue.(int)
		}
		cmd.Flags().IntVarP(ptr, flagName, shorthand, defaultInt, fieldSelector.Description)
	case *metav1.Duration:
		defaultDuration := time.Duration(0)
		if fieldSelector.DefaultValue != nil {
			if dur, ok := fieldSelector.DefaultValue.(metav1.Duration); ok {
				defaultDuration = dur.Duration
			}
		}
		cmd.Flags().DurationVarP(&ptr.Duration, flagName, shorthand, defaultDuration, fieldSelector.Description)
	case *time.Duration:
		defaultDuration := time.Duration(0)
		if fieldSelector.DefaultValue != nil {
			defaultDuration = fieldSelector.DefaultValue.(time.Duration)
		}
		cmd.Flags().DurationVarP(ptr, flagName, shorthand, defaultDuration, fieldSelector.Description)
	}

	// Bind the flag to viper
	if err := m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName)); err != nil {
		// Log error but don't fail - this is not critical
		fmt.Printf("Warning: failed to bind flag %s to viper: %v\n", flagName, err)
	}
}

// generateFlagName generates a user-friendly flag name from a field pointer.
func (m *Manager) generateFlagName(fieldPtr any) string {
	// Check which field this pointer references by comparing addresses
	if fieldPtr == &m.Config.Metadata.Name {
		return "name"
	}
	if fieldPtr == &m.Config.Spec.Distribution {
		return "distribution"
	}
	if fieldPtr == &m.Config.Spec.DistributionConfig {
		return "distribution-config"
	}
	if fieldPtr == &m.Config.Spec.SourceDirectory {
		return "source-directory"
	}
	if fieldPtr == &m.Config.Spec.Connection.Context {
		return "context"
	}
	if fieldPtr == &m.Config.Spec.Connection.Kubeconfig {
		return "kubeconfig"
	}
	if fieldPtr == &m.Config.Spec.Connection.Timeout {
		return "timeout"
	}
	if fieldPtr == &m.Config.Spec.ReconciliationTool {
		return "reconciliation-tool"
	}
	if fieldPtr == &m.Config.Spec.CNI {
		return "cni"
	}
	if fieldPtr == &m.Config.Spec.CSI {
		return "csi"
	}
	if fieldPtr == &m.Config.Spec.IngressController {
		return "ingress-controller"
	}
	if fieldPtr == &m.Config.Spec.GatewayController {
		return "gateway-controller"
	}

	// Fallback to field name detection for other fields
	fieldName := m.getFieldNameFromPointer(fieldPtr, m.Config)
	return strings.ToLower(fieldName)
}

// getFieldNameFromPointer finds the field name for a given pointer.
func (m *Manager) getFieldNameFromPointer(fieldPtr any, rootStruct any) string {
	return m.findFieldName(reflect.ValueOf(fieldPtr), reflect.ValueOf(rootStruct))
}

// findFieldName recursively finds the field name for a pointer.
func (m *Manager) findFieldName(targetPtr, current reflect.Value) string {
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return ""
		}
		current = current.Elem()
	}

	if current.Kind() != reflect.Struct {
		return ""
	}

	typ := current.Type()
	for i := 0; i < current.NumField(); i++ {
		field := current.Field(i)
		fieldType := typ.Field(i)

		if !field.CanAddr() {
			continue
		}

		fieldAddr := field.Addr()
		if fieldAddr.Pointer() == targetPtr.Pointer() {
			return fieldType.Name
		}

		// Recurse into nested structs
		if field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			if result := m.findFieldName(targetPtr, field); result != "" {
				return result
			}
		}
	}

	return ""
}

// generateShorthand generates a shorthand flag from the flag name.
func (m *Manager) generateShorthand(flagName string) string {
	switch flagName {
	case "distribution":
		return "d"
	case "context":
		return "c"
	case "kubeconfig":
		return "k"
	case "timeout":
		return "t"
	case "source-directory":
		return "s"
	case "reconciliation-tool":
		return "r"
	case "distribution-config":
		return ""
	default:
		// For other flags, don't provide shorthand to avoid conflicts
		return ""
	}
}

// getFieldPath returns the JSON path of a field pointer within a struct.
func (m *Manager) getFieldPath(fieldPtr any, rootStruct any) string {
	return m.findFieldPath(reflect.ValueOf(fieldPtr), reflect.ValueOf(rootStruct), "")
}

// findFieldPath recursively finds the path to a field pointer.
func (m *Manager) findFieldPath(targetPtr, current reflect.Value, currentPath string) string {
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return ""
		}
		current = current.Elem()
	}

	if current.Kind() != reflect.Struct {
		return ""
	}

	typ := current.Type()
	for i := 0; i < current.NumField(); i++ {
		field := current.Field(i)
		fieldType := typ.Field(i)

		if !field.CanAddr() {
			continue
		}

		fieldAddr := field.Addr()
		if fieldAddr.Pointer() == targetPtr.Pointer() {
			fieldName := fieldType.Name
			if currentPath == "" {
				return fieldName
			}
			return currentPath + "." + fieldName
		}

		// Recurse into nested structs
		if field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			newPath := fieldType.Name
			if currentPath != "" {
				newPath = currentPath + "." + newPath
			}
			if result := m.findFieldPath(targetPtr, field, newPath); result != "" {
				return result
			}
		}
	}

	return ""
}

// setFieldValue sets a field value using reflection.
func setFieldValue(fieldPtr any, value any) {
	if fieldPtr == nil || value == nil {
		return
	}

	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr || fieldVal.IsNil() {
		return
	}

	fieldVal = fieldVal.Elem()
	valueVal := reflect.ValueOf(value)

	if fieldVal.Type().AssignableTo(valueVal.Type()) {
		fieldVal.Set(valueVal)
	} else if fieldVal.Type().ConvertibleTo(valueVal.Type()) {
		fieldVal.Set(valueVal.Convert(fieldVal.Type()))
	}
}
