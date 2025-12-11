package installer_test

import (
	"testing"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen // Table-driven test with many scenarios
func TestGetInstallTimeout(t *testing.T) {
	t.Parallel()

	t.Run("returns_default_when_cluster_is_nil", func(t *testing.T) {
		t.Parallel()

		timeout := installer.GetInstallTimeout(nil)

		assert.Equal(t, installer.DefaultInstallTimeout, timeout)
	})

	t.Run("returns_default_when_timeout_is_zero", func(t *testing.T) {
		t.Parallel()

		cluster := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Timeout: metav1.Duration{Duration: 0},
				},
			},
		}

		timeout := installer.GetInstallTimeout(cluster)

		assert.Equal(t, installer.DefaultInstallTimeout, timeout)
	})

	t.Run("returns_default_when_timeout_is_negative", func(t *testing.T) {
		t.Parallel()

		cluster := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Timeout: metav1.Duration{Duration: -1 * time.Minute},
				},
			},
		}

		timeout := installer.GetInstallTimeout(cluster)

		assert.Equal(t, installer.DefaultInstallTimeout, timeout)
	})

	t.Run("returns_configured_timeout_when_set", func(t *testing.T) {
		t.Parallel()

		expectedTimeout := 10 * time.Minute
		cluster := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Timeout: metav1.Duration{Duration: expectedTimeout},
				},
			},
		}

		timeout := installer.GetInstallTimeout(cluster)

		assert.Equal(t, expectedTimeout, timeout)
	})

	t.Run("returns_configured_timeout_for_short_duration", func(t *testing.T) {
		t.Parallel()

		expectedTimeout := 30 * time.Second
		cluster := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Timeout: metav1.Duration{Duration: expectedTimeout},
				},
			},
		}

		timeout := installer.GetInstallTimeout(cluster)

		assert.Equal(t, expectedTimeout, timeout)
	})

	t.Run("returns_configured_timeout_for_long_duration", func(t *testing.T) {
		t.Parallel()

		expectedTimeout := 2 * time.Hour
		cluster := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Timeout: metav1.Duration{Duration: expectedTimeout},
				},
			},
		}

		timeout := installer.GetInstallTimeout(cluster)

		assert.Equal(t, expectedTimeout, timeout)
	})
}
