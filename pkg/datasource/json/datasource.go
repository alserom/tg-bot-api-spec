package datasource_json

import (
	"encoding/json"
	"errors"
	"os"

	export_to_json "github.com/alserom/tg-bot-api-spec/pkg/export/json"
	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

type DatasourceJson struct {
	jsonData *export_to_json.JsonData
}

func (dj *DatasourceJson) FillApiSpec(as *spec.ApiSpec) error {
	if dj.jsonData == nil {
		return errors.New("json data missed")
	}

	as.SetVersion(dj.jsonData.Version)
	as.SetReleaseDate(dj.jsonData.ReleaseDate)
	as.SetLink(dj.jsonData.Link)

	ch1 := make(chan error)
	ch2 := make(chan error)

	go func() {
		defer close(ch1)
		addTgTypes(as, dj.jsonData.Types, ch1)
	}()
	go func() {
		defer close(ch2)
		addTgMethods(as, dj.jsonData.Methods, ch2)
	}()

	var lastErr error
	for {
		select {
		case err, ok := <-ch1:
			if !ok {
				ch1 = nil
			}
			if err != nil {
				lastErr = err
			}
		case err, ok := <-ch2:
			if !ok {
				ch2 = nil
			}
			if err != nil {
				lastErr = err
			}
		}
		if ch1 == nil && ch2 == nil {
			break
		}
	}

	return lastErr
}

func NewDatasourceJson(path string) (*DatasourceJson, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = validateInput(content)
	if err != nil {
		return nil, err
	}

	var data export_to_json.JsonData
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	return &DatasourceJson{&data}, nil
}

func addTgTypes(as *spec.ApiSpec, types map[string]export_to_json.TgType, ch chan<- error) {
	deferParent := make(map[string][]*spec.TgTypeSpec)
	deferChild := make(map[string]*spec.TgTypeSpec)

	for _, t := range types {
		tgType, err := spec.NewTgTypeSpec(t.Category, t.Name, t.Link)
		if err != nil {
			ch <- err
			return
		}

		tgType.SetDescription(t.Description)

		if t.Parent != nil {
			parent, exists := as.GetType(string(*t.Parent))
			if !exists {
				deferParent[string(*t.Parent)] = append(deferParent[string(*t.Parent)], tgType)
			} else {
				tgType.SetParent(parent)
			}
		}

		for _, cn := range t.Children {
			child, exists := as.GetType(cn)
			if !exists {
				deferChild[cn] = tgType
			} else {
				tgType.AddChild(child)
			}
		}

		for _, p := range t.Properties {
			tgTypeProperty, err := spec.NewTgTypeSpecProperty(p.Name)
			if err != nil {
				ch <- err
				return
			}

			tgTypeProperty.SetDescription(p.Description)
			tgTypeProperty.SetOptional(p.Optional)

			if p.PredefinedValue != nil {
				value := spec.TgTypeSpecPropertyValue(*p.PredefinedValue)
				tgTypeProperty.SetPredefinedValue(&value)
			}

			for _, dt := range p.Types {
				tgTypeProperty.AddDataType(as.DeclareDataType(dt))
			}

			tgType.AddProperty(tgTypeProperty)
		}

		as.AddType(tgType)
	}

	for parentName, childs := range deferParent {
		parent, exists := as.GetType(parentName)
		if !exists {
			ch <- errors.New("parent type missed")
			return
		}
		for _, child := range childs {
			child.SetParent(parent)
		}
	}

	for childName, parent := range deferChild {
		child, exists := as.GetType(childName)
		if !exists {
			ch <- errors.New("child type missed")
			return
		}
		parent.AddChild(child)
	}
}

func addTgMethods(as *spec.ApiSpec, methods map[string]export_to_json.TgMethod, ch chan<- error) {
	for _, m := range methods {
		tgMethod, err := spec.NewTgMethodSpec(m.Category, m.Name, m.Link)
		if err != nil {
			ch <- err
			return
		}

		tgMethod.SetDescription(m.Description)

		for _, dt := range m.Returns {
			tgMethod.AddReturnType(as.DeclareDataType(dt))
		}

		for _, a := range m.Arguments {
			arg, err := spec.NewTgMethodSpecArgument(a.Name)
			if err != nil {
				ch <- err
				return
			}

			arg.SetDescription(a.Description)
			arg.SetRequired(a.Required)

			for _, dt := range a.Types {
				arg.AddDataType(as.DeclareDataType(dt))
			}

			tgMethod.AddArgument(arg)
		}

		as.AddMethod(tgMethod)
	}
}
