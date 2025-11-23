package registry_test

import (
	"testing"

	registry "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registry"
	"github.com/stretchr/testify/require"
)

func TestCreateOptionsWithDefaults(t *testing.T) {
	t.Parallel()

	opts := registry.CreateOptions{
		Name: "dev-cluster-registry",
		Port: 5000,
	}

	defaulted := opts.WithDefaults()

	require.Equal(t, registry.DefaultEndpointHost, defaulted.Host)
	require.Equal(t, opts.Name, defaulted.VolumeName)
}

func TestCreateOptionsValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		opts    registry.CreateOptions
		wantErr error
	}{
		{
			name: "valid configuration",
			opts: registry.CreateOptions{Name: "ksail", Port: 5000},
		},
		{
			name:    "missing name",
			opts:    registry.CreateOptions{Port: 5000},
			wantErr: registry.ErrNameRequired,
		},
		{
			name:    "invalid port",
			opts:    registry.CreateOptions{Name: "ksail", Port: 70000},
			wantErr: registry.ErrInvalidPort,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.opts.Validate()
			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestCreateOptionsEndpoint(t *testing.T) {
	t.Parallel()

	opts := registry.CreateOptions{Name: "ksail", Port: 5000}
	endpoint := opts.Endpoint()

	require.Equal(t, "localhost:5000", endpoint)
}

func TestStartStopStatusOptionValidation(t *testing.T) {
	t.Parallel()

	require.NoError(t, (registry.StartOptions{Name: "ksail"}).Validate())
	require.ErrorIs(t, (registry.StartOptions{}).Validate(), registry.ErrNameRequired)

	require.NoError(t, (registry.StopOptions{Name: "ksail"}).Validate())
	require.ErrorIs(t, (registry.StopOptions{}).Validate(), registry.ErrNameRequired)

	require.NoError(t, (registry.StatusOptions{Name: "ksail"}).Validate())
	require.ErrorIs(t, (registry.StatusOptions{}).Validate(), registry.ErrNameRequired)
}
