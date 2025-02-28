package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type ValidationErr string

func (e ValidationErr) Error() string {
	return string(e)
}

const (
	ErrEmptyFile        = ValidationErr("YAML file is empty")
	ErrInvalidSchema    = ValidationErr("Schema file is invalid")
	ErrUnmarshalError = ValidationErr("could not unmarshal YAML")
	ErrInvalidYAML = ValidationErr("YAML file is invalid")
)

func ValidateYAML(yamlFile string, schemaFile string) error {
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("could not read YAML file: %w", err)
	}
	var yamlContent map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &yamlContent); err != nil {
		return fmt.Errorf("could not parse YAML: %w", err)
	}

	jsonBytes, err := json.Marshal(yamlContent)
	if err != nil {
		return fmt.Errorf("could not convert YAML to JSON: %w", err)
	}

	schemaJSON, err := loadYAMLSchema(schemaFile)
	if err != nil {
		return fmt.Errorf("could not load schema: %w", err)
	}
	schemaLoader := gojsonschema.NewStringLoader(schemaJSON)

	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return ErrInvalidYAML
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return ErrInvalidYAML
	}

	return nil
}

func loadYAMLSchema(schemaPath string) (string, error) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", ErrEmptyFile
	}
	var schema interface{}
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return "", ErrInvalidSchema
	}
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

