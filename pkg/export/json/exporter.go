package export_to_json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

type JsonExporter struct {
	apiSpec spec.ApiSpec
	data    *JsonData
}

func (je JsonExporter) Export(filename string) error {
	if je.data == nil {
		return errors.New("nothing to export")
	}

	outPath, err := filepath.Abs(strings.TrimSpace(filename))
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(outPath)
	if err == nil && fileInfo.IsDir() {
		outPath += "/spec"
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
			content[path], err = json.Marshal(je.data)
		} else {
			content[path], err = json.MarshalIndent(je.data, "", "    ")
		}

		if err != nil {
			return err
		}
	}

	schemaPath := strings.TrimSuffix(strings.TrimSuffix(paths[0], ".min.json"), ".json") + ".schema.json"
	content[schemaPath] = []byte(strings.TrimSpace(strings.ReplaceAll(Schema, "\t", "    ")))

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

func NewApiSpecExporter(as spec.ApiSpec) (*JsonExporter, error) {
	if err := as.SelfCheck(); err != nil {
		return nil, errors.New("invalid spec: " + err.Error())
	}

	data := &JsonData{
		Version:     as.GetVersion(),
		ReleaseDate: as.GetReleaseDate(),
		Link:        as.GetLink(),
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		fillTypes(data, as.GetTypes())
	}()
	go func() {
		defer wg.Done()
		fillMethods(data, as.GetMethods())
	}()
	wg.Wait()

	return &JsonExporter{as, data}, nil
}

func fillTypes(data *JsonData, types map[string]*spec.TgTypeSpec) {
	data.Types = make(map[string]TgType)

	for _, t := range types {
		tgType := TgType{
			Category:    t.GetCategory(),
			Name:        t.GetName(),
			Link:        t.GetLink(),
			Description: t.GetDescription(),
			Properties:  getProperties(t.GetProperties()),
		}

		parent := t.GetParent()
		if parent != nil {
			v := NilableString(parent.GetName())
			tgType.Parent = &v
		}

		var children []string
		for _, c := range t.GetChildren() {
			children = append(children, c.GetName())
		}
		if len(children) > 1 {
			sort.Strings(children)
		}
		tgType.Children = children

		data.Types[t.GetName()] = tgType
	}
}

func getProperties(props []*spec.TgTypeSpecProperty) []TgTypeProperty {
	var properties []TgTypeProperty
	for _, p := range props {
		ttp := TgTypeProperty{
			Name:        p.GetName(),
			Description: p.GetDescription(),
			Optional:    p.IsOptional(),
		}

		var types []string
		for _, dt := range p.GetDataTypes() {
			types = append(types, dt.GetDefinition())
		}
		if len(types) > 1 {
			sort.Strings(types)
		}
		ttp.Types = types

		value := p.GetPredefinedValue()
		if value != nil {
			v := NilableString(*value)
			ttp.PredefinedValue = &v
		}

		properties = append(properties, ttp)
	}

	if len(properties) > 1 {
		sort.SliceStable(properties, func(i, j int) bool {
			return properties[i].Name < properties[j].Name
		})
	}

	return properties
}

func fillMethods(data *JsonData, methods map[string]*spec.TgMethodSpec) {
	data.Methods = make(map[string]TgMethod)

	for _, m := range methods {
		tgMethod := TgMethod{
			Category:    m.GetCategory(),
			Name:        m.GetName(),
			Link:        m.GetLink(),
			Description: m.GetDescription(),
			Arguments:   getArguments(m.GetArguments()),
		}

		var returns []string
		for _, r := range m.GetReturnTypes() {
			returns = append(returns, r.GetDefinition())
		}
		if len(returns) > 1 {
			sort.Strings(returns)
		}
		tgMethod.Returns = returns

		data.Methods[m.GetName()] = tgMethod
	}
}

func getArguments(args []*spec.TgMethodSpecArgument) []TgMethodArgument {
	var arguments []TgMethodArgument
	for _, a := range args {
		tma := TgMethodArgument{
			Name:        a.GetName(),
			Description: a.GetDescription(),
			Required:    a.IsRequired(),
		}

		var types []string
		for _, dt := range a.GetDataTypes() {
			types = append(types, dt.GetDefinition())
		}
		if len(types) > 1 {
			sort.Strings(types)
		}
		tma.Types = types

		arguments = append(arguments, tma)
	}

	if len(arguments) > 1 {
		sort.SliceStable(arguments, func(i, j int) bool {
			return arguments[i].Name < arguments[j].Name
		})
	}

	return arguments
}
