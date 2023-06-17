package export_to_openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

type OpenapiExporter struct {
	apiSpec spec.ApiSpec
	data    map[string]interface{}
}

func NewOpenapiExporter(as spec.ApiSpec) (*OpenapiExporter, error) {
	if err := as.SelfCheck(); err != nil {
		return nil, errors.New("invalid spec: " + err.Error())
	}

	schemas, err := schemas(&as)
	if err != nil {
		return nil, err
	}

	paths, err := paths(&as)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["openapi"] = "3.1.0"
	data["info"] = info(&as)
	data["externalDocs"] = externalDocs()
	data["servers"] = servers()
	data["paths"] = paths
	data["components"] = map[string]interface{}{"schemas": schemas, "responses": responses()}
	data["security"] = []map[string]interface{}{{}}

	return &OpenapiExporter{as, data}, nil
}

func (oe OpenapiExporter) Export(filename string) error {
	if len(oe.data) == 0 {
		return errors.New("nothing to export")
	}

	outPath, err := filepath.Abs(strings.TrimSpace(filename))
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(outPath)
	if err == nil && fileInfo.IsDir() {
		outPath += "/openapi"
	}

	var paths []string
	if !strings.HasSuffix(outPath, ".json") {
		paths = append(
			paths,
			outPath+".json",
			outPath+".min.json",
		)
	} else {
		paths = append(paths, outPath)
	}

	content := make(map[string][]byte)
	for _, path := range paths {
		if strings.HasSuffix(path, ".min.json") {
			content[path], err = json.Marshal(oe.data)
		} else {
			content[path], err = json.MarshalIndent(oe.data, "", "    ")
		}

		if err != nil {
			return err
		}
	}

	removePathOnFail := ""
	for path, c := range content {
		fmt.Println("saving: " + path)
		err = ioutil.WriteFile(path, c, 0644)
		if err != nil {
			if removePathOnFail != "" {
				os.Remove(removePathOnFail)
			}
			return err
		}
		removePathOnFail = path
	}

	return nil
}
