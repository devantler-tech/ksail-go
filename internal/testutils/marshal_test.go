package testutils_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
)

type testStruct struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

func TestMustMarshal(t *testing.T) {
	t.Parallel()

	t.Run("successful_marshal", func(t *testing.T) {
		t.Parallel()

		marshaller := yamlmarshaller.NewMarshaller[testStruct]()
		data := testStruct{Name: "test", Value: 42}

		result := testutils.MustMarshal(t, marshaller, data)

		// Should contain expected YAML content
		if len(result) == 0 {
			t.Error("Expected non-empty marshaled result")
		}
	})
}

func TestMustUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("successful_unmarshal", func(t *testing.T) {
		t.Parallel()

		marshaller := yamlmarshaller.NewMarshaller[testStruct]()
		yamlData := []byte("name: test\nvalue: 42\n")
		var result testStruct

		testutils.MustUnmarshal(t, marshaller, yamlData, &result)

		// Verify the unmarshaled data
		if result.Name != "test" || result.Value != 42 {
			t.Errorf("Expected {Name: test, Value: 42}, got %+v", result)
		}
	})
}

func TestMustUnmarshalString(t *testing.T) {
	t.Parallel()

	t.Run("successful_unmarshal_string", func(t *testing.T) {
		t.Parallel()

		marshaller := yamlmarshaller.NewMarshaller[testStruct]()
		yamlString := "name: test\nvalue: 42\n"
		var result testStruct

		testutils.MustUnmarshalString(t, marshaller, yamlString, &result)

		// Verify the unmarshaled data
		if result.Name != "test" || result.Value != 42 {
			t.Errorf("Expected {Name: test, Value: 42}, got %+v", result)
		}
	})
}
