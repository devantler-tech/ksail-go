// Package v1alpha1 provides model definitions for a KSail cluster.
package v1alpha1

import (
	"errors"
	"fmt"
	"strings"
)

// --- Errors ---

// ErrInvalidDistribution is returned when an invalid distribution is specified.
var ErrInvalidDistribution = errors.New("invalid distribution")

// ErrInvalidReconciliationTool is returned when an invalid reconciliation tool is specified.
var ErrInvalidReconciliationTool = errors.New("invalid reconciliation tool")

// ErrInvalidContainerEngine is returned when an invalid container engine is specified.
var ErrInvalidContainerEngine = errors.New("invalid container engine")

// --- Getters and Setters ---

// Set for Distribution.
func (d *Distribution) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, dist := range validDistributions() {
		if strings.EqualFold(value, string(dist)) {
			*d = dist

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s, %s)",
		ErrInvalidDistribution, value, DistributionKind, DistributionK3d, DistributionTind)
}

// Set for ReconciliationTool.
func (d *ReconciliationTool) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, tool := range validReconciliationTools() {
		if strings.EqualFold(value, string(tool)) {
			*d = tool

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s, %s)",
		ErrInvalidReconciliationTool, value, ReconciliationToolKubectl, ReconciliationToolFlux, ReconciliationToolArgoCD)
}

// -- pflags --

// String returns the string representation of the Distribution.
func (d *Distribution) String() string {
	return string(*d)
}

// Type returns the type of the Distribution.
func (d *Distribution) Type() string {
	return "Distribution"
}

// String returns the string representation of the ReconciliationTool.
func (d *ReconciliationTool) String() string {
	return string(*d)
}

// Type returns the type of the ReconciliationTool.
func (d *ReconciliationTool) Type() string {
	return "ReconciliationTool"
}
