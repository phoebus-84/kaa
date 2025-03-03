package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	// "fmt"
	"os"
	"testing"
)

const (
	schemaPath      = "./fixtures/Schema.yaml"
	testYAML        = "./fixtures/ValidYAML.yaml"
	testInvalidYaml = "./fixtures/InvalidYAML.yaml"
	validSchemaUrl  = "https://raw.githubusercontent.com/phoebus-84/kaa/refs/heads/main/cmd/fixtures/Schema.yaml"
	validYAMLURL    = "https://raw.githubusercontent.com/phoebus-84/kaa/refs/heads/main/cmd/fixtures/ValidYAML.yaml"
	invalidUrl      = "http://invalid-url"
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
	schema, err := LoadYAMLSchema(schemaPath)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	t.Run("valid YAML", func(t *testing.T) {
		fileToBeValidated, err := LoadYAMLFile(testYAML)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		errV := ValidateYAML(fileToBeValidated, schema)
		if errV != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		buffer := bytes.Buffer{}
		fileToBeValidated, err := LoadYAMLFile(testInvalidYaml)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		errV := ValidateYAML(fileToBeValidated, schema, &buffer)
		got := buffer.String()
		gotArray := strings.Split(got, "\n")
		want := `The document is not valid. Errors:
- version: Does not match pattern '^\d+\.\d+\.\d+$'
- replicas: Must be greater than or equal to 1
- serviceName: Does not match pattern '^[a-zA-Z0-9-]+$'
`
		wantArray := strings.Split(want, "\n")
		if errV == nil {
			t.Errorf("Expected an error, but got nil")
		}

		for _, value := range gotArray {
			if !slices.Contains(wantArray, value) {
				t.Errorf("Expected %q in error message, but it was not found", value)
			}
		}
	})
}

func TestLoadYAMLSchema(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		nonExistentFile := "/path/to/non/existent/file.yaml"
		_, err := LoadYAMLSchema(nonExistentFile)
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
		_, err = LoadYAMLSchema(emptyFile.Name())
		ExpectError(t, err, ErrEmptyFile.Error())
	})

	t.Run("invalid YAML", func(t *testing.T) {
		invalidYAMLFile := CreateTempFile(t, "invalid_yaml", "invalid: yaml: content:")
		_, err := LoadYAMLSchema(invalidYAMLFile.Name())
		ExpectError(t, err, ErrInvalidSchema.Error())
	})

	t.Run("valid schema", func(t *testing.T) {
		jsonSchema, err := LoadYAMLFile(schemaPath)
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
		content, err := LoadYAMLFromURL(validSchemaUrl)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if len(content) == 0 {
			t.Errorf("Expected non-empty data, but got empty")
		}
		jsonSchema, _ := LoadYAMLFile(schemaPath)
		if !reflect.DeepEqual(jsonSchema, content) {
			t.Errorf("Expected data to match schema, but got different data")
		}
	})
}

func ExampleValidateYAML() {
	schema, _ := LoadYAMLSchema(schemaPath)
	fileToBeValidated, _ := LoadYAMLFile(testInvalidYaml)
	buffer := bytes.Buffer{}
	ValidateYAML(fileToBeValidated, schema, &buffer)
	outputSlice := strings.Split(buffer.String(), "\n")
	fmt.Printf("%s", outputSlice[0])
	// Output:
	// The document is not valid. Errors:
}

func ExampleLoadYAMLSchema() {
	schema, _ := LoadYAMLSchema(schemaPath)
	fmt.Printf("%s", schema)
	// Output:
	// {"$schema":"http://json-schema.org/draft-07/schema#","properties":{"replicas":{"description":"Number of instances to run (between 1 and 100).","maximum":100,"minimum":1,"type":"integer"},"serviceName":{"description":"Name of the service (letters, numbers, and hyphens only).","pattern":"^[a-zA-Z0-9-]+$","type":"string"},"version":{"description":"Semantic version number (e.g., 1.2.3).","pattern":"^\\d+\\.\\d+\\.\\d+$","type":"string"}},"required":["serviceName","version","replicas"],"type":"object"}
}

func ExampleLoadYAMLFromURL() {
	yaml, _ := LoadYAMLFromURL(validSchemaUrl)
	fmt.Printf("%s", yaml)
	// Output:
	// {"$schema":"http://json-schema.org/draft-07/schema#","properties":{"replicas":{"description":"Number of instances to run (between 1 and 100).","maximum":100,"minimum":1,"type":"integer"},"serviceName":{"description":"Name of the service (letters, numbers, and hyphens only).","pattern":"^[a-zA-Z0-9-]+$","type":"string"},"version":{"description":"Semantic version number (e.g., 1.2.3).","pattern":"^\\d+\\.\\d+\\.\\d+$","type":"string"}},"required":["serviceName","version","replicas"],"type":"object"}
}

func ExampleLoadYAMLFile() {
	LoadYAMLFile(testYAML)
	// Output:
}
