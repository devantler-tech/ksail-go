package kindprovisioner

import (
	"context"
	"fmt"
	"io"
	"strings"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const kindNetworkName = "kind"

// setupRegistryManager creates a registry manager and extracts registries from Kind config.
// Returns nil if no setup is needed.
func setupRegistryManager(
	kindConfig *v1alpha4.Cluster,
	dockerClient client.APIClient,
	upstreams map[string]string,
) (*dockerclient.RegistryManager, []registries.Info, error) {
	if kindConfig == nil {
		return nil, nil, nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create registry manager: %w", err)
	}

	registriesInfo := extractRegistriesFromKind(kindConfig, upstreams)
	if len(registriesInfo) == 0 {
		return nil, nil, nil
	}

	return registryMgr, registriesInfo, nil
}

// SetupRegistries creates mirror registries based on Kind cluster configuration.
// Registries are created without network attachment first, as the "kind" network
// doesn't exist until after the cluster is created. mirrorSpecs should contain the
// user-supplied mirror definitions so upstream URLs can be preserved when creating
// local proxy registries.
func SetupRegistries(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	clusterName string,
	dockerClient client.APIClient,
	mirrorSpecs []registries.MirrorSpec,
	writer io.Writer,
) error {
	upstreams := registries.BuildUpstreamLookup(mirrorSpecs)

	registryMgr, registriesInfo, err := setupRegistryManager(kindConfig, dockerClient, upstreams)
	if err != nil {
		return err
	}

	if registryMgr == nil {
		return nil
	}

	errSetup := registries.SetupRegistries(ctx, registryMgr, registriesInfo, clusterName, writer)
	if errSetup != nil {
		return fmt.Errorf("failed to setup kind registries: %w", errSetup)
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

	registriesInfo := extractRegistriesFromKind(kindConfig, nil)
	if len(registriesInfo) == 0 {
		return nil
	}

	errConnect := registries.ConnectRegistriesToNetwork(
		ctx,
		dockerClient,
		registriesInfo,
		kindNetworkName,
		writer,
	)
	if errConnect != nil {
		return fmt.Errorf("failed to connect kind registries to network: %w", errConnect)
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
	registryMgr, registriesInfo, err := setupRegistryManager(kindConfig, dockerClient, nil)
	if err != nil {
		return err
	}

	if registryMgr == nil {
		return nil
	}

	errCleanup := registries.CleanupRegistries(
		ctx,
		registryMgr,
		registriesInfo,
		clusterName,
		deleteVolumes,
		nil,
	)
	if errCleanup != nil {
		return fmt.Errorf("failed to cleanup kind registries: %w", errCleanup)
	}

	return nil
}

// ExtractRegistriesFromKindForTesting extracts registry information from Kind configuration.
// This function is exported for testing purposes.
func ExtractRegistriesFromKindForTesting(
	kindConfig *v1alpha4.Cluster,
	upstreams map[string]string,
) []registries.Info {
	return extractRegistriesFromKind(kindConfig, upstreams)
}

// extractRegistriesFromKind is the internal implementation.
func extractRegistriesFromKind(
	kindConfig *v1alpha4.Cluster,
	upstreams map[string]string,
) []registries.Info {
	if kindConfig == nil {
		return nil
	}

	var registryInfos []registries.Info

	seenHosts := make(map[string]bool)
	usedPorts := make(map[int]struct{})
	nextPort := registries.DefaultRegistryPort

	for _, patch := range kindConfig.ContainerdConfigPatches {
		mirrors := parseContainerdConfig(patch)
		if len(mirrors) == 0 {
			continue
		}

		hosts := make([]string, 0, len(mirrors))
		for host := range mirrors {
			hosts = append(hosts, host)
		}

		registries.SortHosts(hosts)

		for _, host := range hosts {
			if seenHosts[host] {
				continue
			}

			seenHosts[host] = true

			endpoints := mirrors[host]
			port := registries.ExtractRegistryPort(endpoints, usedPorts, &nextPort)
			info := registries.BuildRegistryInfo(host, endpoints, port, "kind-", upstreams[host])
			registryInfos = append(registryInfos, info)
		}
	}

	return registryInfos
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
	parts := strings.SplitSeq(arrayContent, ",")
	for part := range parts {
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
