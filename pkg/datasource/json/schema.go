package datasource_json

import (
	"errors"

	export_to_json "github.com/alserom/tg-bot-api-spec/pkg/export/json"

	"github.com/xeipuuv/gojsonschema"
)

func validateInput(data []byte) error {
	schemaLoader := gojsonschema.NewStringLoader(export_to_json.Schema)
	documentLoader := gojsonschema.NewBytesLoader(data)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	errMsg := "fails on schema validation:"
	for _, err := range result.Errors() {
		errMsg += "\n- " + err.String()
	}

	return errors.New(errMsg)
}
