package cmd

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/cobra"
	v "github.com/phoebus-84/Validation"
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



var schemaFile string
var schemaUrl string

var validateCmd = &cobra.Command{
	Use:   "validate [file|url]",
	Short: "Validate a YAML file against a schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		yamlFile := args[0]

		if schemaFile == "" && schemaUrl == "" {
			fmt.Println("Schema file is required. Use --schema to specify a schema file or --schema-url to specify a URL to a schema file.")
			os.Exit(1)
		}
		var schema string
		if schemaFile != "" {
			s, err := v.LoadYAMLSchema(schemaFile)
			if err != nil {
				fmt.Printf("Failed to load schema file: %v\n", err)
				os.Exit(1)
			}
			schema = s
		} else {
			s, err := v.LoadYAMLSchemaFromURL(schemaUrl)
			if err != nil {
				fmt.Printf("Failed to load schema file from URL: %v\n", err)
				os.Exit(1)
			}
			schema = s
		}
		var fileToBeValidated []byte
		if strings.Contains(yamlFile, "http") {
			f, err := v.LoadYAMLFromURL(yamlFile)
			if err != nil {
				fmt.Printf("Failed to load YAML file from URL: %v\n", err)
				os.Exit(1)
			}
			fileToBeValidated = f
		} else {
			f, err := v.LoadYAMLFile(yamlFile)
			if err != nil {
				fmt.Printf("Failed to load YAML file: %v\n", err)
				os.Exit(1)
			}
			fileToBeValidated = f
		}
		if err := v.ValidateYAML(fileToBeValidated, schema); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Println("YAML file is valid!")
		}
	},
}

func init() {
	validateCmd.Flags().StringVarP(&schemaFile, "schema", "s", "", "Path to the JSON Schema file")
	validateCmd.Flags().StringVarP(&schemaUrl, "schema-url", "u", "", "URL to the JSON Schema file")
	rootCmd.AddCommand(validateCmd)
}
