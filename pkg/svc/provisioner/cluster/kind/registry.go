// Package kindprovisioner provides implementations of the Provisioner interface
// for provisioning clusters in different providers.
package kindprovisioner

import (
	"context"
	"fmt"
	"io"
	"strings"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// RegistryInfo holds information about a registry to be created.
type RegistryInfo struct {
	Name     string
	Upstream string
	Port     int
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
	if kindConfig == nil {
		return nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	registries := extractRegistriesFromKind(kindConfig)
	if len(registries) == 0 {
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

		if err := registryMgr.CreateRegistry(ctx, config); err != nil {
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
		containerName := fmt.Sprintf("ksail-registry-%s", reg.Name)

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
	if kindConfig == nil {
		return nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	registries := extractRegistriesFromKind(kindConfig)
	if len(registries) == 0 {
		return nil
	}

	for _, reg := range registries {
		if err := registryMgr.DeleteRegistry(ctx, reg.Name, clusterName, deleteVolumes); err != nil {
			// Log error but don't fail the entire cleanup
			fmt.Printf("Warning: failed to cleanup registry %s: %v\n", reg.Name, err)
		}
	}

	return nil
}

// extractRegistriesFromKind extracts registry information from Kind configuration.
func extractRegistriesFromKind(kindConfig *v1alpha4.Cluster) []RegistryInfo {
	var registries []RegistryInfo
	seenHosts := make(map[string]bool) // Track unique hosts to avoid duplicates
	portOffset := 0

	// Kind uses containerdConfigPatches to configure registry mirrors
	for _, patch := range kindConfig.ContainerdConfigPatches {
		// Parse containerd config to extract registry mirrors
		// Format example:
		// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
		//   endpoint = ["http://localhost:5000"]

		mirrors := parseContainerdConfig(patch)
		for host, endpoints := range mirrors {
			// Skip if we've already processed this host
			if seenHosts[host] {
				continue
			}
			seenHosts[host] = true

			// Generate a simple name from the host (sanitize for container naming)
			name := strings.ReplaceAll(host, ".", "-")
			name = strings.ReplaceAll(name, "/", "-")
			name = strings.ReplaceAll(name, ":", "-")

			// Generate upstream URL from host
			// docker.io -> https://registry-1.docker.io
			// ghcr.io -> https://ghcr.io
			// custom.registry.io:5000 -> https://custom.registry.io:5000
			upstream := generateUpstreamURL(host)

			// Extract port from first endpoint if provided
			port := 5000 + portOffset
			if len(endpoints) > 0 {
				if extractedPort := extractPortFromEndpoint(endpoints[0]); extractedPort > 0 {
					port = extractedPort
				}
			}

			registries = append(registries, RegistryInfo{
				Name:     name,
				Upstream: upstream,
				Port:     port,
			})
			portOffset++
		}
	}

	return registries
}

// parseContainerdConfig parses containerd configuration patches to extract registry mirrors.
// Returns a map of registry host to list of endpoint URLs.
func parseContainerdConfig(patch string) map[string][]string {
	mirrors := make(map[string][]string)
	lines := strings.Split(patch, "\n")

	var currentHost string
	var inEndpointArray bool
	var currentEndpoints []string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Match registry mirror section header
		// Format: [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
		if strings.Contains(line, `registry.mirrors."`) {
			// Save previous host's endpoints if any
			if currentHost != "" && len(currentEndpoints) > 0 {
				mirrors[currentHost] = currentEndpoints
				currentEndpoints = nil
			}

			// Extract new host
			start := strings.Index(line, `mirrors."`)
			if start >= 0 {
				start += len(`mirrors."`)
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					currentHost = line[start : start+end]
					inEndpointArray = false
				}
			}
			continue
		}

		// Match endpoint configuration
		if currentHost != "" && strings.Contains(line, "endpoint") {
			// Handle inline array: endpoint = ["http://..."]
			// Remove all spaces around = for robust parsing
			cleanLine := strings.ReplaceAll(line, " ", "")
			if strings.Contains(cleanLine, `["`) && strings.Contains(cleanLine, `"]`) {
				endpoints := extractEndpointsFromInlineArray(line)
				currentEndpoints = append(currentEndpoints, endpoints...)
				// Save immediately for inline format
				if len(currentEndpoints) > 0 {
					mirrors[currentHost] = currentEndpoints
					currentHost = ""
					currentEndpoints = nil
				}
			} else if strings.Contains(cleanLine, "[") && !strings.Contains(cleanLine, "]") {
				// Start of multiline array: endpoint = [
				inEndpointArray = true
			}
			continue
		}

		// Handle multiline array elements
		if inEndpointArray {
			if strings.Contains(line, "]") {
				// End of array
				inEndpointArray = false
				// Extract any endpoint on the closing line
				if endpoint := extractEndpointFromLine(line); endpoint != "" {
					currentEndpoints = append(currentEndpoints, endpoint)
				}
				// Save collected endpoints
				if currentHost != "" && len(currentEndpoints) > 0 {
					mirrors[currentHost] = currentEndpoints
					currentHost = ""
					currentEndpoints = nil
				}
			} else {
				// Extract endpoint from array element line
				if endpoint := extractEndpointFromLine(line); endpoint != "" {
					currentEndpoints = append(currentEndpoints, endpoint)
				}
			}
		}
	}

	// Save any remaining host endpoints
	if currentHost != "" && len(currentEndpoints) > 0 {
		mirrors[currentHost] = currentEndpoints
	}

	return mirrors
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

// extractQuotedString extracts a string from within quotes.
func extractQuotedString(s string) string {
	s = strings.TrimSpace(s)

	// Find first and last quote
	firstQuote := strings.Index(s, `"`)
	if firstQuote < 0 {
		return ""
	}

	lastQuote := strings.LastIndex(s, `"`)
	if lastQuote <= firstQuote {
		return ""
	}

	return s[firstQuote+1 : lastQuote]
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
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil && port > 0 && port <= 65535 {
			return port
		}
	}

	return 0
}
