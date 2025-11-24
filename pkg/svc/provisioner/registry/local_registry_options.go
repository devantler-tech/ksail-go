package registry

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	// DefaultEndpointHost exposes the registry on the developer workstation only.
	DefaultEndpointHost = "localhost"
	minRegistryPort     = 1
	maxRegistryPort     = 65535
)

// CreateOptions capture the desired shape of a managed OCI registry container.
type CreateOptions struct {
	Name        string
	Host        string
	Port        int
	VolumeName  string
	ClusterName string
}

// WithDefaults applies standard defaults for host bindings and storage metadata.
func (o CreateOptions) WithDefaults() CreateOptions {
	trimmed := o
	trimmed.Name = strings.TrimSpace(trimmed.Name)
	trimmed.Host = strings.TrimSpace(trimmed.Host)
	trimmed.VolumeName = strings.TrimSpace(trimmed.VolumeName)

	if trimmed.Host == "" {
		trimmed.Host = DefaultEndpointHost
	}

	if trimmed.VolumeName == "" {
		trimmed.VolumeName = trimmed.Name
	}

	return trimmed
}

// Validate asserts that the option set is well formed before interacting with the container engine.
func (o CreateOptions) Validate() error {
	opts := o.WithDefaults()

	if opts.Name == "" {
		return ErrNameRequired
	}

	if opts.Port < minRegistryPort || opts.Port > maxRegistryPort {
		return fmt.Errorf("%w: %d", ErrInvalidPort, opts.Port)
	}

	if opts.Host == "" {
		return ErrHostRequired
	}

	return nil
}

// Endpoint returns the host-facing endpoint string (host:port) for the registry.
func (o CreateOptions) Endpoint() string {
	opts := o.WithDefaults()
	if opts.Port <= 0 {
		return opts.Host
	}

	return net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
}

// StartOptions define how a registry instance should be started or connected.
type StartOptions struct {
	Name        string
	NetworkName string
}

// Validate ensures the start options reference a registry container.
func (o StartOptions) Validate() error {
	if strings.TrimSpace(o.Name) == "" {
		return ErrNameRequired
	}

	return nil
}

// StopOptions describe how a registry instance should be stopped and optionally cleaned up.
type StopOptions struct {
	Name         string
	ClusterName  string
	NetworkName  string
	DeleteVolume bool
}

// Validate ensures the stop options reference a registry container.
func (o StopOptions) Validate() error {
	if strings.TrimSpace(o.Name) == "" {
		return ErrNameRequired
	}

	return nil
}

// StatusOptions identify the registry instance whose status should be inspected.
type StatusOptions struct {
	Name string
}

// Validate ensures the status query references a registry container.
func (o StatusOptions) Validate() error {
	if strings.TrimSpace(o.Name) == "" {
		return ErrNameRequired
	}

	return nil
}
