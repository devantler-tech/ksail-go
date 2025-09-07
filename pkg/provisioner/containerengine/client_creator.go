// Package containerengine provides client creation functionality for container engines.
package containerengine

import "github.com/docker/docker/client"

// ClientCreator is a function type for creating container engine clients.
type ClientCreator func() (client.APIClient, error)

// ClientCreators holds optional client creator functions for dependency injection.
// This provides a more robust API than positional parameters, making the interface
// self-documenting and eliminating ordering concerns.
type ClientCreators struct {
	// Docker specifies a custom Docker client creator.
	// If nil, GetDockerClient will be used.
	Docker ClientCreator
	
	// PodmanUser specifies a custom Podman user client creator.
	// If nil, GetPodmanUserClient will be used.
	PodmanUser ClientCreator
	
	// PodmanSystem specifies a custom Podman system client creator.
	// If nil, GetPodmanSystemClient will be used.
	PodmanSystem ClientCreator
}