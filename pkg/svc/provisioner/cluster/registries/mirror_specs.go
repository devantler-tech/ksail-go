package registries

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// MirrorSpec represents a parsed mirror registry specification entry.
type MirrorSpec struct {
	Host   string
	Remote string
}

// MirrorEntry contains the normalized data required to create a registry mirror.
type MirrorEntry struct {
	Host          string
	SanitizedName string
	ContainerName string
	Endpoint      string
	Port          int
}

// ParseMirrorSpecs converts raw mirror specification strings into structured specs.
// Invalid entries (missing host or remote) are ignored.
func ParseMirrorSpecs(specs []string) []MirrorSpec {
	parsed := make([]MirrorSpec, 0, len(specs))

	for _, raw := range specs {
		host, remote, ok := splitMirrorSpec(raw)
		if !ok {
			continue
		}

		host = strings.TrimSpace(host)

		remote = strings.TrimSpace(remote)
		if host == "" || remote == "" {
			continue
		}

		parsed = append(parsed, MirrorSpec{
			Host:   host,
			Remote: remote,
		})
	}

	return parsed
}

// BuildMirrorEntries converts mirror specs into registry entries using the provided prefix.
// Prefix should exclude the trailing hyphen (e.g., "kind", "k3d").
// Existing hosts are skipped, and the allocations update the provided maps.
func BuildMirrorEntries(
	specs []MirrorSpec,
	containerPrefix string,
	existingHosts map[string]struct{},
	usedPorts map[int]struct{},
	nextPort *int,
) []MirrorEntry {
	targetHosts := existingHosts
	if targetHosts == nil {
		targetHosts = map[string]struct{}{}
	}

	allocatedPorts := usedPorts
	if allocatedPorts == nil {
		allocatedPorts = map[int]struct{}{}
	}

	entries := make([]MirrorEntry, 0, len(specs))

	for _, spec := range specs {
		host := strings.TrimSpace(spec.Host)
		if host == "" {
			continue
		}

		if _, exists := targetHosts[host]; exists {
			continue
		}

		sanitized := SanitizeHostIdentifier(host)
		if sanitized == "" {
			continue
		}

		port := AllocatePort(nextPort, allocatedPorts)
		containerName := fmt.Sprintf("%s-%s", containerPrefix, sanitized)
		endpoint := "http://" + net.JoinHostPort(containerName, strconv.Itoa(port))

		entries = append(entries, MirrorEntry{
			Host:          host,
			SanitizedName: sanitized,
			ContainerName: containerName,
			Endpoint:      endpoint,
			Port:          port,
		})

		targetHosts[host] = struct{}{}
	}

	return entries
}

// BuildUpstreamLookup returns a map of registry host to user-specified upstream URL.
func BuildUpstreamLookup(specs []MirrorSpec) map[string]string {
	if len(specs) == 0 {
		return nil
	}

	lookup := make(map[string]string, len(specs))

	for _, spec := range specs {
		host := strings.TrimSpace(spec.Host)

		remote := strings.TrimSpace(spec.Remote)
		if host == "" || remote == "" {
			continue
		}

		lookup[host] = remote
	}

	if len(lookup) == 0 {
		return nil
	}

	return lookup
}

// AllocatePort returns the next available port and updates the tracking map.
func AllocatePort(nextPort *int, usedPorts map[int]struct{}) int {
	if nextPort == nil {
		value := DefaultRegistryPort
		nextPort = &value
	}

	if usedPorts == nil {
		usedPorts = map[int]struct{}{}
	}

	port := *nextPort
	if port <= 0 {
		port = DefaultRegistryPort
	}

	for {
		if _, exists := usedPorts[port]; !exists {
			usedPorts[port] = struct{}{}
			*nextPort = port + 1

			return port
		}

		port++
	}
}

// KindPatch renders a containerd mirror patch for the provided entry.
func KindPatch(entry MirrorEntry) string {
	return fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
  endpoint = ["%s"]`, entry.Host, entry.Endpoint)
}

func splitMirrorSpec(spec string) (string, string, bool) {
	idx := strings.Index(spec, "=")
	if idx <= 0 || idx == len(spec)-1 {
		return "", "", false
	}

	return spec[:idx], spec[idx+1:], true
}
