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
	ErrEmptyFile      = ValidationErr("YAML file is empty")
	ErrInvalidSchema  = ValidationErr("Schema file is invalid")
	ErrUnmarshalError = ValidationErr("could not unmarshal YAML")
	ErrInvalidYAML    = ValidationErr("YAML file is invalid")
)

func ValidateYAML(yamlFile string, schemaFile string) error {
	yamlContent, err := LoadYAMLFile(yamlFile)
	if err != nil {
		return err
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
	schema, err := LoadYAMLFile(schemaPath)
	if err == ErrUnmarshalError {
		return "", ErrInvalidSchema
	}
	if err != nil {
		return "", err
	}
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func LoadYAMLFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrEmptyFile
	}
	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, ErrUnmarshalError
	}
	return content, nil
}

func LoadYAMLFromURL(url string) ([]byte, error) {
	return nil, nil
}
