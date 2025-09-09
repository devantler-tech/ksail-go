package k8s_test

import (
	"testing"
	"time"

	k8sutils "github.com/devantler-tech/ksail-go/internal/utils/k8s"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewEmptyObjectMeta(t *testing.T) {
	t.Parallel()

	meta := k8sutils.NewEmptyObjectMeta()

	// Assert - all string fields should be empty
	assert.Empty(t, meta.Name)
	assert.Empty(t, meta.GenerateName)
	assert.Empty(t, meta.Namespace)
	assert.Empty(t, meta.SelfLink) //nolint:staticcheck // Testing legacy field
	assert.Empty(t, string(meta.UID))
	assert.Empty(t, meta.ResourceVersion)

	// Assert - numeric fields should be zero
	assert.Equal(t, int64(0), meta.Generation)

	// Assert - time fields should be empty
	assert.Equal(t, metav1.Time{Time: time.Time{}}, meta.CreationTimestamp)

	// Assert - pointer fields should be nil
	assert.Nil(t, meta.DeletionTimestamp)
	assert.Nil(t, meta.DeletionGracePeriodSeconds)
	assert.Nil(t, meta.Labels)
	assert.Nil(t, meta.Annotations)
	assert.Nil(t, meta.OwnerReferences)
	assert.Nil(t, meta.Finalizers)
	assert.Nil(t, meta.ManagedFields)
}
