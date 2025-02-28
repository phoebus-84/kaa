package cmd

import (
	"encoding/json"
	// "fmt"
	"os"
	"testing"
)

const (
  schema = "./fixtures/Schema.yaml"
  testYAML = "./fixtures/ValidYAML.yaml"
	testInvalidYaml = "./fixtures/InvalidYAML.yaml"
)

func ExpectError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if err.Error() != msg {
		t.Errorf("Expected error message %q, but got %q", msg, err.Error())
	}
}

func CreateTempFile(t *testing.T, prefix string, content string) *os.File {
	t.Helper()
	file, err := os.CreateTemp("", prefix)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestValidateYAML(t *testing.T) {
	t.Run("valid YAML", func(t *testing.T) {
		err := ValidateYAML(testYAML, schema)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		err := ValidateYAML(testInvalidYaml, schema)
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
	})
}

func ExampleValidateYAML() {
	// schemaFile := CreateTempFile(t, "schema", testYAMLSchema)
	// yamlFile := CreateTempFile(t, "valid_yaml", testYAML)
	// err := ValidateYAML(yamlFile.Name(), schemaFile.Name())
	// if err != nil {
	//   fmt.Println(err)
	// }
	// Output:
}

func TestLoadYAMLSchema(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		nonExistentFile := "/path/to/non/existent/file.yaml"
		_, err := loadYAMLSchema(nonExistentFile)
		if err == nil {
			t.Errorf("Expected an error when loading non-existent file, but got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		emptyFile, err := os.CreateTemp("", "empty_schema")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(emptyFile.Name())
		defer emptyFile.Close()
		_, err = loadYAMLSchema(emptyFile.Name())
		ExpectError(t, err, ErrEmptyFile.Error())
	})

	t.Run("invalid YAML", func(t *testing.T) {
		invalidYAMLFile := CreateTempFile(t, "invalid_yaml", "invalid: yaml: content:")
		_, err := loadYAMLSchema(invalidYAMLFile.Name())
		ExpectError(t, err, ErrInvalidSchema.Error())
	})

	t.Run("valid schema", func(t *testing.T) {
		jsonSchema, err := loadYAMLSchema(schema)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		var jsonContent map[string]interface{}
		err = json.Unmarshal([]byte(jsonSchema), &jsonContent)
		if err != nil {
			t.Fatalf("Expected valid JSON, but got error: %v", err)
		}
		expectedKeys := []string{"$schema", "type", "required", "properties"}
		for _, key := range expectedKeys {
			if _, ok := jsonContent[key]; !ok {
				t.Errorf("Expected key %q in converted JSON, but it was not found", key)
			}
		}
	})
}

func TestLoadYAMLFromUrl(t *testing.T) {
	t.Run("non-existent URL", func(t *testing.T) {
		_, err := LoadYAMLFromURL("http://non-existent-url.com")
		if err == nil {
			t.Errorf("Expected an error when loading from non-existent URL, but got nil")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := LoadYAMLFromURL("http://invalid-url")
		if err == nil {
			t.Errorf("Expected an error when loading from invalid URL, but got nil")
		}
	})

	t.Run("valid URL", func(t *testing.T) {
		// Skip this test as it requires a valid URL
		t.Skip("Skipping test that requires a valid URL")
	})
}
