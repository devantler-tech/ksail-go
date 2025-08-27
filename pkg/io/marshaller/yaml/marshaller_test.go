package yamlmarshaller_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sample model used for tests.
type sample struct {
	Name   string   `json:"name"           yaml:"name"`
	Count  int      `json:"count"          yaml:"count"`
	Active bool     `json:"active"         yaml:"active"`
	Tags   []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func TestMarshal_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	mar := yamlmarshaller.NewMarshaller[sample]()
	want := sample{
		Name:   "app",
		Count:  3,
		Active: true,
		Tags:   []string{"dev", "test"},
	}

	// Act
	out, err := mar.Marshal(want)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, out)

	// Round-trip to ensure content encodes the same data
	var got sample

	testutils.MustUnmarshalString[sample](t, mar, out, &got)
	assert.Equal(t, want, got)
}

func TestMarshalString_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	mar := yamlmarshaller.NewMarshaller[sample]()
	input := sample{
		Name:   "app",
		Count:  3,
		Active: true,
		Tags:   []string{"dev", "test"},
	}

	// Act
	out := testutils.MustMarshal(t, mar, input)
	// Some yaml libs may preserve struct field names; accept either lowercase (from tags) or field name casing.
	testutils.AssertStringContainsOneOf(t, out, "name: app", "Name: app")
	testutils.AssertStringContainsOneOf(t, out, "count: 3", "Count: 3")
	testutils.AssertStringContainsOneOf(t, out, "active: true", "Active: true")
	testutils.AssertStringContains(t, out, "- dev", "- test")
}

func TestUnmarshal_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	mar := yamlmarshaller.NewMarshaller[sample]()
	data := []byte("" +
		"name: app\n" +
		"count: 3\n" +
		"active: true\n" +
		"tags:\n" +
		"- dev\n" +
		"- test\n",
	)
	want := sample{
		Name:   "app",
		Count:  3,
		Active: true,
		Tags:   []string{"dev", "test"},
	}

	// Act
	var got sample

	testutils.MustUnmarshal[sample](t, mar, data, &got)
	assert.Equal(t, want, got)
}

func TestUnmarshalString_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	mar := yamlmarshaller.NewMarshaller[sample]()
	data := "" +
		"name: app\n" +
		"count: 3\n" +
		"active: true\n" +
		"tags:\n" +
		"- dev\n" +
		"- test\n"
	want := sample{
		Name:   "app",
		Count:  3,
		Active: true,
		Tags:   []string{"dev", "test"},
	}

	// Act
	var got sample

	testutils.MustUnmarshalString[sample](t, mar, data, &got)
	assert.Equal(t, want, got)
}

func TestMarshal_Error_UnsupportedType(t *testing.T) {
	t.Parallel()

	// Arrange: a type that cannot be marshaled (contains a func field)
	type bad struct {
		F func()
	}

	mar := yamlmarshaller.NewMarshaller[bad]()
	input := bad{F: func() {}}

	// Act
	yamlText, err := mar.Marshal(input)

	// Assert
	require.Error(t, err)
	assert.Empty(t, yamlText)
	assert.ErrorContains(t, err, "failed to marshal YAML")
}

func TestUnmarshal_Error_UnsupportedType(t *testing.T) {
	t.Parallel()

	// Arrange: a type that cannot be unmarshaled (contains a func field)
	type bad struct {
		F func()
	}

	mar := yamlmarshaller.NewMarshaller[bad]()
	input := bad{F: func() {}}

	// Act
	err := mar.Unmarshal([]byte("F: !!js/function 'function() {}'"), &input)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to unmarshal YAML")
}

func TestUnmarshalString_Error_UnsupportedType(t *testing.T) {
	t.Parallel()

	// Arrange: a type that cannot be unmarshaled (contains a func field)
	type bad struct {
		F func()
	}

	mar := yamlmarshaller.NewMarshaller[bad]()
	input := bad{F: func() {}}

	// Act
	err := mar.UnmarshalString("F: !!js/function 'function() {}'", &input)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to unmarshal YAML")
}
