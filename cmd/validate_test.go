package cmd

import (
	"encoding/json"
	// "fmt"
	"os"
	"testing"
)

const (
	testYAMLSchema = `
$schema: "http://json-schema.org/draft-07/schema#"
type: object
required:
  - serviceName
  - version
  - replicas
properties:
  serviceName:
    type: string
    pattern: "^[a-zA-Z0-9-]+$"
    description: "Name of the service (letters, numbers, and hyphens only)."
  version:
    type: string
    pattern: "^\\d+\\.\\d+\\.\\d+$"
    description: "Semantic version number (e.g., 1.2.3)."
  replicas:
    type: integer
    minimum: 1
    maximum: 100
    description: "Number of instances to run (between 1 and 100)."
`
	testYAML = `
serviceName: "user-service"
version: "1.2.3"
replicas: 3
`
	testInvalidYaml = `
serviceName: "user_service"
version: "1.2"
replicas: 0
`
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
		schemaFile := CreateTempFile(t, "schema", testYAMLSchema)
		yamlFile := CreateTempFile(t, "valid_yaml", testYAML)
		err := ValidateYAML(yamlFile.Name(), schemaFile.Name())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		schemaFile := CreateTempFile(t, "schema", testYAMLSchema)
		invalidYAMLFile := CreateTempFile(t, "invalid_yaml", testInvalidYaml)
		err := ValidateYAML(invalidYAMLFile.Name(), schemaFile.Name())
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
		validSchemaFile := CreateTempFile(t, "valid_schema", testYAMLSchema)
		jsonSchema, err := loadYAMLSchema(validSchemaFile.Name())
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
