package docker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	errPingFailed        = errors.New("ping failed")
	errShouldNotRun      = errors.New("should not run")
	errDockerUnavailable = errors.New("docker unavailable")
	errPodmanPingFailed  = errors.New("podman ping failed")
	errCallFailed        = errors.New("call failed")
)

func TestNewContainerEngine(t *testing.T) {
	t.Parallel()

	t.Run("returns error when client is nil", func(t *testing.T) {
		t.Parallel()

		engine, err := docker.NewContainerEngine(nil)
		if err == nil {
			t.Fatalf("expected error when client is nil")
		}

		if !errors.Is(err, docker.ErrAPIClientNil) {
			t.Fatalf("expected ErrAPIClientNil, got %v", err)
		}

		if engine != nil {
			t.Fatalf("expected no engine, got %v", engine)
		}
	})

	t.Run("wraps provided client", func(t *testing.T) {
		t.Parallel()

		mockClient := docker.NewMockAPIClient(t)

		engine, err := docker.NewContainerEngine(mockClient)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if engine == nil {
			t.Fatalf("expected engine, got nil")
		}

		if engine.Client != mockClient {
			t.Fatalf("engine did not retain provided client")
		}
	})
}

func TestContainerEngineCheckReady(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pingErr error
		ready   bool
	}{
		{name: "ready when ping succeeds", pingErr: nil, ready: true},
		{name: "not ready when ping fails", pingErr: errPingFailed, ready: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockClient := docker.NewMockAPIClient(t)
			mockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, testCase.pingErr)

			engine, err := docker.NewContainerEngine(mockClient)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ready, checkErr := engine.CheckReady(context.Background())
			if ready != testCase.ready {
				t.Fatalf("unexpected ready status: got %v want %v", ready, testCase.ready)
			}

			if testCase.pingErr == nil && checkErr != nil {
				t.Fatalf("unexpected error: %v", checkErr)
			}

			if testCase.pingErr != nil && checkErr == nil {
				t.Fatalf("expected error when ping fails")
			}
		})
	}
}

func TestGetAutoDetectedClientReturnsFirstReady(t *testing.T) {
	t.Parallel()

	secondCalled := false

	engine, err := docker.GetAutoDetectedClient(
		readyCreator(t, dockerVersion()),
		func() (client.APIClient, error) {
			secondCalled = true

			return nil, errShouldNotRun
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine == nil {
		t.Fatalf("expected engine, got nil")
	}

	if secondCalled {
		t.Fatalf("unexpected invocation of fallback creator")
	}

	if engine.GetName() != "Docker" {
		t.Fatalf("unexpected engine name: %s", engine.GetName())
	}
}

func TestGetAutoDetectedClientFallsBackWhenFirstCreatorNotReady(t *testing.T) {
	t.Parallel()

	engine, err := docker.GetAutoDetectedClient(
		notReadyCreator(t, errPingFailed),
		readyCreator(t, podmanVersion()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine == nil {
		t.Fatalf("expected engine, got nil")
	}

	if engine.GetName() != "Podman" {
		t.Fatalf("unexpected engine name: %s", engine.GetName())
	}
}

func TestGetAutoDetectedClientReturnsErrorWhenNoCreatorsSucceed(t *testing.T) {
	t.Parallel()

	engine, err := docker.GetAutoDetectedClient(
		failingCreator(errDockerUnavailable),
		notReadyCreator(t, errPodmanPingFailed),
	)

	if !errors.Is(err, docker.ErrNoContainerEngine) {
		t.Fatalf("expected ErrNoContainerEngine, got %v", err)
	}

	if engine != nil {
		t.Fatalf("expected nil engine when detection fails")
	}
}

func TestContainerEngineGetName(t *testing.T) {
	t.Parallel()

	t.Run("detects docker from platform", func(t *testing.T) {
		t.Parallel()

		engine := engineWithVersion(t, dockerVersion(), nil)

		if engine.GetName() != "Docker" {
			t.Fatalf("unexpected engine name: %s", engine.GetName())
		}
	})

	t.Run("detects podman from version string", func(t *testing.T) {
		t.Parallel()

		version := versionWithPlatform("", "5.0.0-PodMan")
		engine := engineWithVersion(t, version, nil)

		if engine.GetName() != "Podman" {
			t.Fatalf("unexpected engine name: %s", engine.GetName())
		}
	})

	t.Run("returns unknown when detection fails", func(t *testing.T) {
		t.Parallel()

		engine := engineWithVersion(t, emptyVersion(), nil)

		if engine.GetName() != "Unknown" {
			t.Fatalf("expected Unknown, got %s", engine.GetName())
		}
	})

	t.Run("returns unknown when server version fails", func(t *testing.T) {
		t.Parallel()

		engine := engineWithVersion(t, types.Version{}, errCallFailed)

		if engine.GetName() != "Unknown" {
			t.Fatalf("expected Unknown, got %s", engine.GetName())
		}
	})
}

func TestGetDockerClient(t *testing.T) {
	t.Parallel()

	client, err := docker.GetDockerClient()
	if err != nil {
		if client != nil {
			t.Fatalf("expected nil client on error, got %v", client)
		}

		return
	}

	if client == nil {
		t.Fatalf("expected client when no error returned")
	}
}

func TestGetPodmanUserClient(t *testing.T) {
	t.Parallel()

	client, err := docker.GetPodmanUserClient()
	if err != nil {
		if client != nil {
			t.Fatalf("expected nil client on error, got %v", client)
		}

		return
	}

	if client == nil {
		t.Fatalf("expected client when no error returned")
	}
}

func TestGetPodmanSystemClient(t *testing.T) {
	t.Parallel()

	client, err := docker.GetPodmanSystemClient()
	if err != nil {
		if client != nil {
			t.Fatalf("expected nil client on error, got %v", client)
		}

		return
	}

	if client == nil {
		t.Fatalf("expected client when no error returned")
	}
}

func readyCreator(t *testing.T, version types.Version) docker.ClientCreator {
	t.Helper()

	mockClient := docker.NewMockAPIClient(t)
	mockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	mockClient.EXPECT().ServerVersion(context.Background()).Return(version, nil)

	return func() (client.APIClient, error) {
		return mockClient, nil
	}
}

func notReadyCreator(t *testing.T, pingErr error) docker.ClientCreator {
	t.Helper()

	mockClient := docker.NewMockAPIClient(t)
	mockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, pingErr)

	return func() (client.APIClient, error) {
		return mockClient, nil
	}
}

func failingCreator(err error) docker.ClientCreator {
	return func() (client.APIClient, error) {
		return nil, err
	}
}

func engineWithVersion(
	t *testing.T,
	version types.Version,
	versionErr error,
) *docker.ContainerEngine {
	t.Helper()

	mockClient := docker.NewMockAPIClient(t)
	mockClient.EXPECT().ServerVersion(context.Background()).Return(version, versionErr)

	engine, err := docker.NewContainerEngine(mockClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return engine
}

func versionWithPlatform(name, version string) types.Version {
	return types.Version{
		Platform: struct{ Name string }{Name: name},
		Version:  version,
	}
}

func dockerVersion() types.Version {
	return versionWithPlatform("Docker Engine - Community", "24.0.0")
}

func podmanVersion() types.Version {
	return versionWithPlatform("Podman Engine", "5.0.0")
}

func emptyVersion() types.Version {
	return versionWithPlatform("", "")
}
