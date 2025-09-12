package testutils_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalFailer(t *testing.T) {
	t.Parallel()

	t.Run("marshal_always_fails", func(t *testing.T) {
		t.Parallel()

		// Create a marshal failer
		failer := testutils.MarshalFailer[testConfig]{
			Marshaller: yamlmarshaller.NewMarshaller[testConfig](),
		}

		config := testConfig{Name: "test"}

		// Marshal should always fail
		result, err := failer.Marshal(config)

		require.Error(t, err)
		assert.Equal(t, testutils.ErrMarshalFailed, err)
		assert.Empty(t, result)
	})
}

func TestErrMarshalFailed(t *testing.T) {
	t.Parallel()

	t.Run("error_message", func(t *testing.T) {
		t.Parallel()

		// Verify the error message
		assert.Equal(t, "marshal failed", testutils.ErrMarshalFailed.Error())
	})
}
