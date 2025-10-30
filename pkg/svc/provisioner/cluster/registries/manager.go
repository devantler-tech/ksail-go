// Package registries contains helpers for managing shared mirror registry state across
// different provisioners. Functions here are used by Kind and K3d implementations
// to create, connect, and clean up registry containers consistently.
package registries

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
)

// Info describes a registry mirror that should be managed for a cluster.
type Info struct {
	Host     string
	Name     string
	Upstream string
	Port     int
	Volume   string
}

// DefaultRegistryPort defines the default container registry port inside the container.
const DefaultRegistryPort = 5000

const expectedEndpointParts = 2

// SetupRegistries ensures that the provided registries exist. Any newly created
// registries are cleaned up if a later creation fails.
func SetupRegistries(
	ctx context.Context,
	registryMgr *dockerclient.RegistryManager,
	registries []Info,
	clusterName string,
	writer io.Writer,
) error {
	if registryMgr == nil || len(registries) == 0 {
		return nil
	}

	existingRegistries, err := collectExistingRegistryNames(ctx, registryMgr)
	if err != nil {
		return err
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "creating mirror registries",
		Writer:  writer,
	})

	createdRegistries := make([]Info, 0, len(registries))

	for _, reg := range registries {
		created, createErr := ensureRegistry(
			ctx,
			registryMgr,
			clusterName,
			reg,
			writer,
			existingRegistries,
		)
		if createErr != nil {
			cleanupCreatedRegistries(ctx, registryMgr, createdRegistries, clusterName, writer)

			return createErr
		}

		if created {
			createdRegistries = append(createdRegistries, reg)
		}
	}

	return nil
}

func collectExistingRegistryNames(
	ctx context.Context,
	registryMgr *dockerclient.RegistryManager,
) (map[string]struct{}, error) {
	existingRegistries := make(map[string]struct{})

	current, listErr := registryMgr.ListRegistries(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("failed to list existing registries: %w", listErr)
	}

	for _, name := range current {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		existingRegistries[trimmed] = struct{}{}
	}

	return existingRegistries, nil
}

func ensureRegistry(
	ctx context.Context,
	registryMgr *dockerclient.RegistryManager,
	clusterName string,
	reg Info,
	writer io.Writer,
	existing map[string]struct{},
) (bool, error) {
	notify.WriteMessage(notify.Message{
		Type: notify.ActivityType,
		Content: fmt.Sprintf(
			"creating mirror for %s on http://localhost:%d",
			reg.Upstream,
			reg.Port,
		),
		Writer: writer,
	})

	config := dockerclient.RegistryConfig{
		Name:        reg.Name,
		Port:        reg.Port,
		UpstreamURL: reg.Upstream,
		ClusterName: clusterName,
		NetworkName: "",
		VolumeName:  reg.Volume,
	}

	_, alreadyExists := existing[reg.Name]

	err := registryMgr.CreateRegistry(ctx, config)
	if err != nil {
		return false, fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
	}

	if alreadyExists {
		return false, nil
	}

	existing[reg.Name] = struct{}{}

	return true, nil
}

func cleanupCreatedRegistries(
	ctx context.Context,
	registryMgr *dockerclient.RegistryManager,
	created []Info,
	clusterName string,
	writer io.Writer,
) {
	for i := len(created) - 1; i >= 0; i-- {
		reg := created[i]

		err := registryMgr.DeleteRegistry(ctx, reg.Name, clusterName, false)
		if err != nil {
			notify.WriteMessage(notify.Message{
				Type: notify.WarningType,
				Content: fmt.Sprintf(
					"cleanup warning: failed to delete registry %s: %v",
					reg.Name,
					err,
				),
				Writer: writer,
			})
		}
	}
}

// ConnectRegistriesToNetwork attaches each registry container to the provided network.
// Any connection failures are logged as warnings but do not abort the operation.
func ConnectRegistriesToNetwork(
	ctx context.Context,
	dockerClient client.APIClient,
	registries []Info,
	networkName string,
	writer io.Writer,
) error {
	if dockerClient == nil || len(registries) == 0 || strings.TrimSpace(networkName) == "" {
		return nil
	}

	for _, reg := range registries {
		containerName := reg.Name
		if strings.TrimSpace(containerName) == "" {
			continue
		}

		err := dockerClient.NetworkConnect(ctx, networkName, containerName, nil)
		if err != nil {
			notify.WriteMessage(notify.Message{
				Type: notify.WarningType,
				Content: fmt.Sprintf(
					"failed to connect registry %s to %s network: %v",
					containerName,
					networkName,
					err,
				),
				Writer: writer,
			})
		}
	}

	return nil
}

// CleanupRegistries removes the provided registries. Errors are logged as warnings.
func CleanupRegistries(
	ctx context.Context,
	registryMgr *dockerclient.RegistryManager,
	registries []Info,
	clusterName string,
	deleteVolumes bool,
	warningWriter io.Writer,
) error {
	if registryMgr == nil || len(registries) == 0 {
		return nil
	}

	writer := warningWriter
	if writer == nil {
		writer = os.Stderr
	}

	for _, reg := range registries {
		err := registryMgr.DeleteRegistry(ctx, reg.Name, clusterName, deleteVolumes)
		if err != nil {
			_, _ = fmt.Fprintf(
				writer,
				"Warning: failed to cleanup registry %s: %v\n",
				reg.Name,
				err,
			)
		}
	}

	return nil
}

// SanitizeHostIdentifier converts a registry host string into a filesystem-safe identifier.
func SanitizeHostIdentifier(host string) string {
	sanitized := strings.ReplaceAll(host, ".", "-")
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")

	return sanitized
}

// GenerateVolumeName returns a deterministic Docker volume name for the registry.
func GenerateVolumeName(host string) string {
	return SanitizeHostIdentifier(host)
}

// GenerateUpstreamURL attempts to derive the upstream registry URL from the host name.
func GenerateUpstreamURL(host string) string {
	if host == "docker.io" {
		return "https://registry-1.docker.io"
	}

	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}

	return "https://" + host
}

// ExtractRegistryPort determines a unique host port to expose for the given endpoints.
func ExtractRegistryPort(endpoints []string, usedPorts map[int]struct{}, nextPort *int) int {
	if nextPort == nil {
		defaultPort := DefaultRegistryPort
		nextPort = &defaultPort
	}

	if candidate := firstAvailableEndpointPort(endpoints, usedPorts, nextPort); candidate > 0 {
		return candidate
	}

	port := *nextPort
	for {
		if port <= 0 {
			port = DefaultRegistryPort
		}

		if _, exists := usedPorts[port]; !exists {
			break
		}

		port++
	}

	usedPorts[port] = struct{}{}
	*nextPort = port + 1

	return port
}

func firstAvailableEndpointPort(
	endpoints []string,
	usedPorts map[int]struct{},
	nextPort *int,
) int {
	if len(endpoints) == 0 {
		return 0
	}

	extracted := ExtractPortFromEndpoint(endpoints[0])
	if extracted <= 0 {
		return 0
	}

	if _, exists := usedPorts[extracted]; exists {
		return 0
	}

	usedPorts[extracted] = struct{}{}
	if extracted >= *nextPort {
		*nextPort = extracted + 1
	}

	return extracted
}

// ExtractPortFromEndpoint extracts the port from an endpoint URL. Returns 0 if not found.
func ExtractPortFromEndpoint(endpoint string) int {
	lastColon := strings.LastIndex(endpoint, ":")
	if lastColon < 0 {
		return 0
	}

	portStr := endpoint[lastColon+1:]
	if slashIdx := strings.Index(portStr, "/"); slashIdx >= 0 {
		portStr = portStr[:slashIdx]
	}

	var port int

	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil || port <= 0 || port > 65535 {
		return 0
	}

	return port
}

// ResolveRegistryName determines the registry container name from endpoints or falls back to prefix + host.
func ResolveRegistryName(host string, endpoints []string, prefix string) string {
	for _, endpoint := range endpoints {
		if name := ExtractNameFromEndpoint(endpoint); name != "" && !isLocalEndpointName(name) {
			return name
		}
	}

	return BuildRegistryName(prefix, host)
}

func isLocalEndpointName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "localhost" || lower == "0.0.0.0" {
		return true
	}

	return strings.HasPrefix(lower, "127.")
}

// ExtractNameFromEndpoint extracts the hostname portion from an endpoint URL.
func ExtractNameFromEndpoint(endpoint string) string {
	parts := strings.Split(endpoint, "//")
	if len(parts) != expectedEndpointParts {
		return ""
	}

	hostPort := strings.Split(parts[1], ":")
	if len(hostPort) == 0 {
		return ""
	}

	return hostPort[0]
}

// BuildRegistryName constructs a registry container name from prefix and host.
func BuildRegistryName(prefix, host string) string {
	sanitized := SanitizeHostIdentifier(host)

	return prefix + sanitized
}

// BuildRegistryInfo creates an Info populated with derived fields using the supplied prefix for container names.
func BuildRegistryInfo(
	host string,
	endpoints []string,
	port int,
	prefix string,
	upstreamOverride string,
) Info {
	name := ResolveRegistryName(host, endpoints, prefix)

	upstream := strings.TrimSpace(upstreamOverride)
	if upstream == "" {
		upstream = GenerateUpstreamURL(host)
	}

	volume := GenerateVolumeName(host)

	return Info{
		Host:     host,
		Name:     name,
		Upstream: upstream,
		Port:     port,
		Volume:   volume,
	}
}

// SortHosts deterministically sorts registry hostnames.
func SortHosts(hosts []string) {
	sort.Strings(hosts)
}
