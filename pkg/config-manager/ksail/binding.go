// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains field selector binding functionality for automatic CLI flag creation.
package ksail

import (
	"reflect"
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

// addFlagFromField adds a CLI flag for a specific field using type assertion and reflection.
//
//nolint:gocognit,nestif,cyclop,funlen // Complex type switching is necessary for type-safe flag binding
func (m *Manager) addFlagFromField(
	cmd *cobra.Command,
	fieldSelector FieldSelector[v1alpha1.Cluster],
) {
	fieldPtr := fieldSelector.Selector(m.Config)
	if fieldPtr == nil {
		return
	}

	flagName := m.GenerateFlagName(fieldPtr)
	shorthand := m.GenerateShorthand(flagName)

	// Check if the field implements pflag.Value interface first (for enum types)
	if pflagValue, isPflagValue := fieldPtr.(interface {
		Set(value string) error
		String() string
		Type() string
	}); isPflagValue {
		// Set default value if provided
		if fieldSelector.DefaultValue != nil {
			m.setPflagValueDefault(pflagValue, fieldSelector.DefaultValue)
		}

		// Use VarP for pflag.Value types to preserve type information
		if shorthand != "" {
			cmd.Flags().VarP(pflagValue, flagName, shorthand, fieldSelector.Description)
		} else {
			cmd.Flags().Var(pflagValue, flagName, fieldSelector.Description)
		}
	} else {
		// Handle standard types that don't implement pflag.Value
		switch ptr := fieldPtr.(type) {
		case *string:
			defaultStr := ""

			if fieldSelector.DefaultValue != nil {
				if str, ok := fieldSelector.DefaultValue.(string); ok {
					defaultStr = str
				}
			}

			cmd.Flags().StringVarP(ptr, flagName, shorthand, defaultStr, fieldSelector.Description)
		case *bool:
			defaultBool := false

			if fieldSelector.DefaultValue != nil {
				if b, ok := fieldSelector.DefaultValue.(bool); ok {
					defaultBool = b
				}
			}

			cmd.Flags().BoolVarP(ptr, flagName, shorthand, defaultBool, fieldSelector.Description)
		case *int:
			defaultInt := 0

			if fieldSelector.DefaultValue != nil {
				if i, ok := fieldSelector.DefaultValue.(int); ok {
					defaultInt = i
				}
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
				if dur, ok := fieldSelector.DefaultValue.(time.Duration); ok {
					defaultDuration = dur
				}
			}

			cmd.Flags().DurationVarP(ptr, flagName, shorthand, defaultDuration, fieldSelector.Description)
		}
	}

	// Bind the flag to viper (ignoring error for non-critical binding)
	_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
}

// GenerateFlagName generates a user-friendly flag name from a field pointer.
//
//nolint:cyclop // Field mapping requires many conditional checks for type safety
func (m *Manager) GenerateFlagName(fieldPtr any) string {
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

	// For unknown fields, return a simple fallback
	return "unknown"
}

// GenerateShorthand generates a shorthand flag from the flag name.
func (m *Manager) GenerateShorthand(flagName string) string {
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

// setPflagValueDefault sets the default value for a pflag.Value.
func (m *Manager) setPflagValueDefault(pflagValue interface {
	Set(value string) error
	String() string
	Type() string
}, defaultValue any,
) {
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
