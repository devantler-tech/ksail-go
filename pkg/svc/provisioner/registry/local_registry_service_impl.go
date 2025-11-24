package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

// Config controls how the registry service interacts with the container engine.
type Config struct {
	DockerClient    client.APIClient
	RegistryManager Backend
}

type service struct {
	docker   client.APIClient
	registry Backend
	ctrl     *Controller
}

// NewService constructs a registry lifecycle manager backed by the Docker API.
func NewService(cfg Config) (Service, error) {
	if cfg.DockerClient == nil {
		return nil, fmt.Errorf("docker client is required")
	}

	backend := cfg.RegistryManager
	if backend == nil {
		manager, err := dockerclient.NewRegistryManager(cfg.DockerClient)
		if err != nil {
			return nil, fmt.Errorf("create registry manager: %w", err)
		}

		backend = manager
	}

	controller, err := NewController(backend)
	if err != nil {
		return nil, err
	}

	return &service{
		docker:   cfg.DockerClient,
		registry: backend,
		ctrl:     controller,
	}, nil
}

func (s *service) Create(ctx context.Context, opts CreateOptions) (v1alpha1.OCIRegistry, error) {
	if err := opts.Validate(); err != nil {
		return v1alpha1.NewOCIRegistry(), err
	}

	resolved := opts.WithDefaults()

	if _, err := s.ctrl.EnsureOne(ctx, resolved.toRegistryInfo(), resolved.ClusterName, io.Discard); err != nil {
		model := buildRegistryModel(resolved.Name, resolved.Endpoint(), int32(resolved.Port), resolved.VolumeName)
		model.Status = v1alpha1.OCIRegistryStatusError
		model.LastError = err.Error()

		return model, fmt.Errorf("create registry: %w", err)
	}

	return s.Status(ctx, StatusOptions{Name: resolved.Name})
}

func (s *service) Start(ctx context.Context, opts StartOptions) (v1alpha1.OCIRegistry, error) {
	if err := opts.Validate(); err != nil {
		return v1alpha1.NewOCIRegistry(), err
	}

	summary, err := s.findRegistryContainer(ctx, opts.Name)
	if err != nil {
		return v1alpha1.NewOCIRegistry(), err
	}

	if !strings.EqualFold(summary.State, "running") {
		if err := s.docker.ContainerStart(ctx, summary.ID, container.StartOptions{}); err != nil {
			return v1alpha1.NewOCIRegistry(), fmt.Errorf("start registry container: %w", err)
		}
	}

	if networkName := strings.TrimSpace(opts.NetworkName); networkName != "" {
		if err := s.ensureNetworkAttachment(ctx, summary.ID, networkName); err != nil {
			return v1alpha1.NewOCIRegistry(), err
		}
	}

	return s.Status(ctx, StatusOptions{Name: opts.Name})
}

func (s *service) Stop(ctx context.Context, opts StopOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	if opts.DeleteVolume {
		return s.ctrl.CleanupOne(ctx, Info{Name: opts.Name}, opts.ClusterName, true, opts.NetworkName, io.Discard)
	}

	summary, err := s.findRegistryContainer(ctx, opts.Name)
	if err != nil {
		if errors.Is(err, dockerclient.ErrRegistryNotFound) {
			return nil
		}

		return err
	}

	if strings.EqualFold(summary.State, "running") {
		if err := s.docker.ContainerStop(ctx, summary.ID, container.StopOptions{}); err != nil {
			return fmt.Errorf("stop registry container: %w", err)
		}
	}

	if networkName := strings.TrimSpace(opts.NetworkName); networkName != "" {
		if err := s.docker.NetworkDisconnect(ctx, networkName, summary.ID, true); err != nil {
			if !client.IsErrNotFound(err) {
				return fmt.Errorf("disconnect registry from network %s: %w", networkName, err)
			}
		}
	}

	return nil
}

func (o CreateOptions) toRegistryInfo() Info {
	opts := o.WithDefaults()

	return Info{
		Host:   opts.Host,
		Name:   opts.Name,
		Port:   opts.Port,
		Volume: opts.VolumeName,
	}
}

func (s *service) Status(ctx context.Context, opts StatusOptions) (v1alpha1.OCIRegistry, error) {
	if err := opts.Validate(); err != nil {
		return v1alpha1.NewOCIRegistry(), err
	}

	model := v1alpha1.NewOCIRegistry()
	model.Name = opts.Name
	model.DataPath = dockerclient.RegistryDataPath

	containers, err := s.listRegistryContainers(ctx, opts.Name)
	if err != nil {
		return model, err
	}

	if len(containers) == 0 {
		return model, nil
	}

	summary := containers[0]
	model.VolumeName = extractVolumeName(summary, opts.Name)

	hostPort := resolveHostPort(summary)
	if hostPort == 0 {
		port, portErr := s.registry.GetRegistryPort(ctx, opts.Name)
		if portErr != nil {
			switch {
			case errors.Is(portErr, dockerclient.ErrRegistryNotFound),
				errors.Is(portErr, dockerclient.ErrRegistryPortNotFound):
				// No additional metadata available.
			default:
				return model, fmt.Errorf("get registry port: %w", portErr)
			}
		} else {
			hostPort = port
		}
	}

	if hostPort > 0 {
		model.Port = int32(hostPort)
		host := resolveEndpointHost(summary)
		model.Endpoint = net.JoinHostPort(host, strconv.Itoa(hostPort))
	}

	model.Status, model.LastError = mapContainerState(summary.State, summary.Status)

	return model, nil
}

func (s *service) listRegistryContainers(ctx context.Context, name string) ([]container.Summary, error) {
	filtersArgs := filters.NewArgs()
	filtersArgs.Add("name", name)
	filtersArgs.Add("ancestor", dockerclient.RegistryImageName)
	filtersArgs.Add("label", fmt.Sprintf("%s=%s", dockerclient.RegistryLabelKey, name))

	containers, err := s.docker.ContainerList(ctx, container.ListOptions{All: true, Filters: filtersArgs})
	if err != nil {
		return nil, fmt.Errorf("list registry containers: %w", err)
	}

	return containers, nil
}

func (s *service) findRegistryContainer(ctx context.Context, name string) (container.Summary, error) {
	containers, err := s.listRegistryContainers(ctx, name)
	if err != nil {
		return container.Summary{}, err
	}

	if len(containers) == 0 {
		return container.Summary{}, dockerclient.ErrRegistryNotFound
	}

	return containers[0], nil
}

func (s *service) ensureNetworkAttachment(ctx context.Context, containerID, networkName string) error {
	inspect, err := s.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("inspect registry container: %w", err)
	}

	if inspect.NetworkSettings != nil {
		if _, exists := inspect.NetworkSettings.Networks[networkName]; exists {
			return nil
		}
	}

	if err := s.docker.NetworkConnect(ctx, networkName, containerID, &network.EndpointSettings{}); err != nil {
		return fmt.Errorf("connect registry to network %s: %w", networkName, err)
	}

	return nil
}

func buildRegistryModel(name, endpoint string, port int32, volumeName string) v1alpha1.OCIRegistry {
	model := v1alpha1.NewOCIRegistry()
	model.Name = name
	model.Endpoint = endpoint
	model.Port = port
	model.VolumeName = volumeName
	model.DataPath = dockerclient.RegistryDataPath
	model.Status = v1alpha1.OCIRegistryStatusProvisioning

	return model
}

func extractVolumeName(summary container.Summary, fallback string) string {
	for _, mountPoint := range summary.Mounts {
		if mountPoint.Type == mount.TypeVolume && strings.TrimSpace(mountPoint.Name) != "" {
			return mountPoint.Name
		}
	}

	if sanitized := dockerclient.NormalizeVolumeName(fallback); sanitized != "" {
		return sanitized
	}

	return strings.TrimSpace(fallback)
}

func resolveHostPort(summary container.Summary) int {
	for _, port := range summary.Ports {
		if int(port.PrivatePort) == dockerclient.DefaultRegistryPort {
			if port.PublicPort > 0 {
				return int(port.PublicPort)
			}
		}
	}

	return 0
}

func resolveEndpointHost(summary container.Summary) string {
	for _, port := range summary.Ports {
		if int(port.PrivatePort) != dockerclient.DefaultRegistryPort {
			continue
		}

		host := strings.TrimSpace(port.IP)
		if host != "" && host != "0.0.0.0" {
			return host
		}
	}

	return DefaultEndpointHost
}

func mapContainerState(state, status string) (v1alpha1.OCIRegistryStatus, string) {
	lower := strings.ToLower(strings.TrimSpace(state))

	switch lower {
	case "running":
		return v1alpha1.OCIRegistryStatusRunning, ""
	case "created", "starting", "restarting":
		return v1alpha1.OCIRegistryStatusProvisioning, ""
	case "paused", "exited", "dead", "removing", "error":
		return v1alpha1.OCIRegistryStatusError, strings.TrimSpace(status)
	default:
		if lower == "" {
			return v1alpha1.OCIRegistryStatusNotProvisioned, ""
		}

		return v1alpha1.OCIRegistryStatusError, strings.TrimSpace(status)
	}
}
