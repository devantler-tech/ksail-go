package kindprovisioner

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
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const (
	// Default port for registry services.
	defaultRegistryPort = 5000
	// Expected parts count when splitting endpoint URLs and host:port strings.
	expectedPartCount = 2
)

// RegistryInfo holds information about a registry to be created.
type RegistryInfo struct {
	Name     string
	Upstream string
	Port     int
}

// setupRegistryManager creates a registry manager and extracts registries from Kind config.
// Returns nil if no setup is needed.
func setupRegistryManager(
	kindConfig *v1alpha4.Cluster,
	dockerClient client.APIClient,
) (*dockerclient.RegistryManager, []RegistryInfo, error) {
	if kindConfig == nil {
		return nil, nil, nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create registry manager: %w", err)
	}

	registries := extractRegistriesFromKind(kindConfig)
	if len(registries) == 0 {
		return nil, nil, nil
	}

	return registryMgr, registries, nil
}

// SetupRegistries creates mirror registries based on Kind cluster configuration.
// Registries are created without network attachment first, as the "kind" network
// doesn't exist until after the cluster is created.
func SetupRegistries(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	clusterName string,
	dockerClient client.APIClient,
	writer io.Writer,
) error {
	// Setup registry manager and extract registries
	registryMgr, registries, err := setupRegistryManager(kindConfig, dockerClient)
	if err != nil {
		return err
	}

	if registryMgr == nil {
		return nil
	}

	// Display activity message for creating registries
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "creating mirror registries",
		Writer:  writer,
	})

	for _, reg := range registries {
		// Display activity message for each registry
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
			NetworkName: "", // Don't attach to network yet - it doesn't exist
		}

		err := registryMgr.CreateRegistry(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
		}
	}

	return nil
}

// ConnectRegistriesToNetwork connects existing registries to the Kind network.
// This should be called after the Kind cluster is created and the "kind" network exists.
func ConnectRegistriesToNetwork(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	dockerClient client.APIClient,
	writer io.Writer,
) error {
	if kindConfig == nil {
		return nil
	}

	registries := extractRegistriesFromKind(kindConfig)
	if len(registries) == 0 {
		return nil
	}

	// Connect each registry to the kind network
	for _, reg := range registries {
		containerName := "ksail-registry-" + reg.Name

		err := dockerClient.NetworkConnect(ctx, "kind", containerName, nil)
		if err != nil {
			// Log warning but don't fail - registry can still work via localhost
			notify.WriteMessage(notify.Message{
				Type: notify.WarningType,
				Content: fmt.Sprintf(
					"failed to connect registry %s to kind network: %v",
					reg.Name,
					err,
				),
				Writer: writer,
			})
		}
	}

	return nil
}

// CleanupRegistries removes registries that are no longer in use.
func CleanupRegistries(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	clusterName string,
	dockerClient client.APIClient,
	deleteVolumes bool,
) error {
	// Setup registry manager and extract registries
	registryMgr, registries, err := setupRegistryManager(kindConfig, dockerClient)
	if err != nil {
		return err
	}

	if registryMgr == nil {
		return nil
	}

	for _, reg := range registries {
		err := registryMgr.DeleteRegistry(ctx, reg.Name, clusterName, deleteVolumes)
		if err != nil {
			// Log error but don't fail the entire cleanup
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Warning: failed to cleanup registry %s: %v\n",
				reg.Name,
				err,
			)
		}
	}

	return nil
}

// ExtractRegistriesFromKindForTesting extracts registry information from Kind configuration.
// This function is exported for testing purposes.
func ExtractRegistriesFromKindForTesting(kindConfig *v1alpha4.Cluster) []RegistryInfo {
	return extractRegistriesFromKind(kindConfig)
}

// extractRegistriesFromKind is the internal implementation.
func extractRegistriesFromKind(kindConfig *v1alpha4.Cluster) []RegistryInfo {
	var registries []RegistryInfo

	seenHosts := make(map[string]bool) // Track unique hosts to avoid duplicates
	portOffset := 0

	// Kind uses containerdConfigPatches to configure registry mirrors
	for _, patch := range kindConfig.ContainerdConfigPatches {
		mirrors := parseContainerdConfig(patch)

		// Sort hosts for deterministic order
		hosts := make([]string, 0, len(mirrors))
		for host := range mirrors {
			hosts = append(hosts, host)
		}

		sort.Strings(hosts)

		for _, host := range hosts {
			// Skip duplicates - don't increment portOffset for duplicates
			if seenHosts[host] {
				continue
			}

			seenHosts[host] = true

			// Build registry info from host and endpoints
			info := buildRegistryInfo(host, mirrors[host], portOffset)
			registries = append(registries, info)
			portOffset++ // Only increment for actually added registries
		}
	}

	return registries
}

// buildRegistryInfo constructs a RegistryInfo from host and endpoints.
func buildRegistryInfo(host string, endpoints []string, portOffset int) RegistryInfo {
	name := extractRegistryName(host, endpoints)
	port := extractRegistryPort(endpoints, portOffset)
	upstream := generateUpstreamURL(host)

	return RegistryInfo{
		Name:     name,
		Upstream: upstream,
		Port:     port,
	}
}

// extractRegistryName extracts or generates a registry name from host and endpoints.
func extractRegistryName(host string, endpoints []string) string {
	if len(endpoints) == 0 {
		return generateNameFromHost(host)
	}

	endpoint := endpoints[0]

	// Try to extract name from endpoint like "http://kind-docker-io:5000"
	if strings.HasPrefix(endpoint, "http://kind-") || strings.HasPrefix(endpoint, "https://kind-") {
		if name := extractNameFromEndpoint(endpoint); name != "" {
			return name
		}
	}

	return generateNameFromHost(host)
}

// extractNameFromEndpoint extracts the registry name from an endpoint URL.
func extractNameFromEndpoint(endpoint string) string {
	parts := strings.Split(endpoint, "//")
	if len(parts) != expectedPartCount {
		return ""
	}

	hostPort := strings.Split(parts[1], ":")
	if len(hostPort) >= 1 {
		return hostPort[0] // Return full distribution-prefixed name like "kind-docker-io"
	}

	return ""
}

// generateNameFromHost generates a registry name from a host.
func generateNameFromHost(host string) string {
	// Sanitize host for container naming
	simpleName := strings.ReplaceAll(host, ".", "-")
	simpleName = strings.ReplaceAll(simpleName, "/", "-")
	simpleName = strings.ReplaceAll(simpleName, ":", "-")

	return "kind-" + simpleName
}

// extractRegistryPort determines the registry port from endpoints or uses default with offset.
func extractRegistryPort(endpoints []string, portOffset int) int {
	if len(endpoints) == 0 {
		return defaultRegistryPort + portOffset
	}

	if extractedPort := extractPortFromEndpoint(endpoints[0]); extractedPort > 0 {
		return extractedPort
	}

	return defaultRegistryPort + portOffset
}

// ParseContainerdConfigForTesting parses containerd configuration patches to extract registry mirrors.
// Returns a map of registry host to list of endpoint URLs.
// This function is exported for testing purposes.
func ParseContainerdConfigForTesting(patch string) map[string][]string {
	return parseContainerdConfig(patch)
}

// parseContainerdConfig is the internal implementation.
func parseContainerdConfig(patch string) map[string][]string {
	mirrors := make(map[string][]string)
	lines := strings.Split(patch, "\n")

	var currentHost string

	var inEndpointArray bool

	var currentEndpoints []string

	for i := range lines {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Process registry mirror section headers
		if strings.Contains(line, `registry.mirrors."`) {
			currentHost = processHostMirrorsLine(
				line,
				currentHost,
				currentEndpoints,
				mirrors,
			)
			currentEndpoints = nil
			inEndpointArray = false

			continue
		}

		// Process endpoint configuration lines
		if currentHost != "" && strings.Contains(line, "endpoint") {
			currentHost, currentEndpoints, inEndpointArray = processEndpointLine(
				line,
				currentHost,
				currentEndpoints,
				mirrors,
			)

			continue
		}

		// Process multiline array elements
		if inEndpointArray {
			currentHost, currentEndpoints, inEndpointArray = processMultilineArrayLine(
				line,
				currentHost,
				currentEndpoints,
				mirrors,
			)
		}
	}

	// Save any remaining host endpoints
	saveHostEndpoints(currentHost, currentEndpoints, mirrors)

	return mirrors
}

// processHostMirrorsLine processes a registry mirrors section header line.
// Returns updated currentHost.
func processHostMirrorsLine(
	line string,
	currentHost string,
	currentEndpoints []string,
	mirrors map[string][]string,
) string {
	// Save previous host's endpoints if any
	saveHostEndpoints(currentHost, currentEndpoints, mirrors)

	// Extract new host from line like: [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
	start := strings.Index(line, `mirrors."`)
	if start < 0 {
		return ""
	}

	start += len(`mirrors."`)

	end := strings.Index(line[start:], `"`)
	if end <= 0 {
		return ""
	}

	newHost := line[start : start+end]

	// Skip if this host is already in mirrors (preserve first occurrence)
	if _, exists := mirrors[newHost]; exists {
		return ""
	}

	return newHost
}

// processEndpointLine processes an endpoint configuration line.
// Returns updated currentHost, currentEndpoints, and inEndpointArray.
func processEndpointLine(
	line string,
	currentHost string,
	currentEndpoints []string,
	mirrors map[string][]string,
) (string, []string, bool) {
	// Check for inline array format: endpoint = ["http://..."]
	cleanLine := strings.ReplaceAll(line, " ", "")
	if strings.Contains(cleanLine, `["`) && strings.Contains(cleanLine, `"]`) {
		endpoints := extractEndpointsFromInlineArray(line)
		currentEndpoints = append(currentEndpoints, endpoints...)
		// Save immediately for inline format
		saveHostEndpoints(currentHost, currentEndpoints, mirrors)

		return "", nil, false
	}

	// Check for multiline array start: endpoint = [
	if strings.Contains(cleanLine, "[") && !strings.Contains(cleanLine, "]") {
		return currentHost, currentEndpoints, true
	}

	return currentHost, currentEndpoints, false
}

// processMultilineArrayLine processes a line within a multiline endpoint array.
// Returns updated currentHost, currentEndpoints, and inEndpointArray.
func processMultilineArrayLine(
	line string,
	currentHost string,
	currentEndpoints []string,
	mirrors map[string][]string,
) (string, []string, bool) {
	if strings.Contains(line, "]") {
		// End of array - extract any endpoint on the closing line
		if endpoint := extractEndpointFromLine(line); endpoint != "" {
			currentEndpoints = append(currentEndpoints, endpoint)
		}
		// Save collected endpoints
		saveHostEndpoints(currentHost, currentEndpoints, mirrors)

		return "", nil, false
	}

	// Extract endpoint from array element line
	if endpoint := extractEndpointFromLine(line); endpoint != "" {
		currentEndpoints = append(currentEndpoints, endpoint)
	}

	return currentHost, currentEndpoints, true
}

// saveHostEndpoints saves endpoints for a host to the mirrors map.
func saveHostEndpoints(host string, endpoints []string, mirrors map[string][]string) {
	if host != "" && len(endpoints) > 0 {
		mirrors[host] = endpoints
	}
}

// extractEndpointsFromInlineArray extracts all endpoints from an inline array format.
// Example: endpoint = ["http://localhost:5000", "http://localhost:5001"]
func extractEndpointsFromInlineArray(line string) []string {
	var endpoints []string

	// Find the array content between [ and ]
	start := strings.Index(line, "[")

	end := strings.LastIndex(line, "]")
	if start < 0 || end < 0 || start >= end {
		return endpoints
	}

	arrayContent := line[start+1 : end]

	// Split by comma and extract quoted strings
	parts := strings.Split(arrayContent, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if endpoint := extractQuotedString(part); endpoint != "" {
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

// extractEndpointFromLine extracts an endpoint URL from a line.
// Handles quoted strings with surrounding whitespace.
func extractEndpointFromLine(line string) string {
	line = strings.TrimSpace(line)
	// Remove trailing comma if present
	line = strings.TrimSuffix(line, ",")

	return extractQuotedString(line)
}

// ExtractQuotedStringForTesting extracts a string from within quotes.
// This function is exported for testing purposes.
func ExtractQuotedStringForTesting(str string) string {
	return extractQuotedString(str)
}

// extractQuotedString is the internal implementation.
func extractQuotedString(str string) string {
	str = strings.TrimSpace(str)

	// Find first and last quote
	firstQuote := strings.Index(str, `"`)
	if firstQuote < 0 {
		return ""
	}

	lastQuote := strings.LastIndex(str, `"`)
	if lastQuote <= firstQuote {
		return ""
	}

	return str[firstQuote+1 : lastQuote]
}

// generateUpstreamURL generates the upstream registry URL from the host.
// docker.io -> https://registry-1.docker.io
// ghcr.io -> https://ghcr.io
// quay.io -> https://quay.io
// custom.registry.io:5000 -> https://custom.registry.io:5000
func generateUpstreamURL(host string) string {
	// Special case for docker.io - the actual registry is registry-1.docker.io
	if host == "docker.io" {
		return "https://registry-1.docker.io"
	}

	// For other registries, use https:// + host
	return "https://" + host
}

// extractPortFromEndpoint extracts the port number from an endpoint URL.
// Returns 0 if no port can be extracted.
func extractPortFromEndpoint(endpoint string) int {
	// Look for port after last colon
	// Example: http://localhost:5000 -> 5000
	lastColon := strings.LastIndex(endpoint, ":")
	if lastColon < 0 {
		return 0
	}

	// Make sure this is actually a port (not part of http://)
	if lastColon < len(endpoint)-1 {
		portStr := endpoint[lastColon+1:]
		// Remove any trailing slash or path
		if slashIdx := strings.Index(portStr, "/"); slashIdx >= 0 {
			portStr = portStr[:slashIdx]
		}

		var port int

		_, err := fmt.Sscanf(portStr, "%d", &port)
		if err == nil && port > 0 && port <= 65535 {
			return port
		}
	}

	return 0
}
