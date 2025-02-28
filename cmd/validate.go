package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
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

func ValidateYAML(fileToBeValidated []byte, schema string, args ...io.Writer) error {
	var w io.Writer
	if len(args) < 1 {
		w = os.Stdout
	} else {
		w = args[0]
	}
	// fileToBeVAlidated, err := LoadYAMLFile(yamlFile)
	// if err != nil {
	// 	return err
	// }
	// schema, err := LoadYAMLSchema(schemaFile)
	// if err != nil {
	// 	return err
	// }
	schemaLoader := gojsonschema.NewStringLoader(schema)

	documentLoader := gojsonschema.NewBytesLoader(fileToBeValidated)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return ErrInvalidYAML
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			fmt.Fprintf(w, "- %s\n", desc)
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

func LoadYAMLSchemaFromURL(url string) (string, error) {
	jsonBytes, err := LoadYAMLFromURL(url)
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

var schemaFile string

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a YAML file against a schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		yamlFile := args[0]

		if schemaFile == "" {
			fmt.Println("Schema file is required. Use --schema to specify.")
			os.Exit(1)
		}
		fileToBeValidated, err := LoadYAMLFile(yamlFile)
		if err != nil {
			fmt.Printf("Failed to load YAML file: %v\n", err)
			os.Exit(1)
		}
		schema, err := LoadYAMLSchema(schemaFile)
		if err != nil {
			fmt.Printf("Failed to load schema file: %v\n", err)
			os.Exit(1)
		}

		if err := ValidateYAML(fileToBeValidated, schema); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Println("YAML file is valid!")
		}
	},
}

func init() {
	validateCmd.Flags().StringVarP(&schemaFile, "schema", "s", "", "Path to the JSON Schema file")
	rootCmd.AddCommand(validateCmd)
}
