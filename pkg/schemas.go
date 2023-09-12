package pkg

import (
	_ "embed"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schemas/page.schema.json
var pageSchemaString string
var PageSchema = jsonschema.MustCompileString("", pageSchemaString)

//go:embed schemas/manifest.schema.json
var manifestSchemaString string
var ManifestSchema = jsonschema.MustCompileString("", manifestSchemaString)

func ValidatePage(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := PageSchema.Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateManifest(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := ManifestSchema.Validate(v); err != nil {
		return err
	}
	return nil
}
