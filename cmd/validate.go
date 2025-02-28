package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	ErrJSONMarshal    = ValidationErr("Cannot marshal YAML into JSON")
)

func ValidateYAML(yamlFile string, schemaFile string) error {
	fileToBeVAlidated, err := LoadYAMLFile(yamlFile)
	if err != nil {
		return err
	}
	schema, err := LoadYAMLSchema(schemaFile)
	if err != nil {
		return err
	}
	schemaLoader := gojsonschema.NewStringLoader(schema)

	documentLoader := gojsonschema.NewBytesLoader(fileToBeVAlidated)

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

func LoadYAMLFile(path string) ([]byte, error) {
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
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		return nil, ErrJSONMarshal
	}
	return jsonBytes, nil
}

func LoadYAMLSchema(path string) (string, error) {
	jsonBytes, err := LoadYAMLFile(path)
	switch err {
	case ErrEmptyFile:
		return "", ErrEmptyFile
	case ErrUnmarshalError:
		return "", ErrInvalidSchema
	case ErrJSONMarshal:
		return "", ErrJSONMarshal
	case nil:
		return string(jsonBytes), nil
	default:
		return "", err
	}
}

func LoadYAMLFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to GET from URL %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status %s from URL %q", resp.Status, url)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from URL %q: %w", url, err)
	}
	if len(data) == 0 {
		return nil, ErrEmptyFile
	}
	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, ErrUnmarshalError
	}
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		return nil, ErrJSONMarshal
	}
	return jsonBytes, nil
}
