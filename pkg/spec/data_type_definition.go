package spec

import (
	"strings"
)

type DataTypeDefinition interface {
	GetDefinition() string
}

type abstractDataType struct {
	definition string
}

func (dt abstractDataType) GetDefinition() string {
	return dt.definition
}

type ScalarDataType struct {
	abstractDataType
}

type ObjectDataType struct {
	abstractDataType
	ref *TgTypeSpec
}

func (o *ObjectDataType) setRef(ref *TgTypeSpec) {
	o.ref = ref
}

func (o ObjectDataType) GetRef() *TgTypeSpec {
	return o.ref
}

type ArrayDataType struct {
	abstractDataType
	elementAnyOf []DataTypeDefinition
}

func (a ArrayDataType) GetElementDataTypes() []DataTypeDefinition {
	return a.elementAnyOf
}

func newDataTypeDefinition(definition string, as *ApiSpec) DataTypeDefinition {
	switch {
	case strings.HasPrefix(definition, "array"):
		elementAnyOf := make([]DataTypeDefinition, 0)
		if strings.HasPrefix(definition, "array<") && strings.HasSuffix(definition, ">") {
			elementDefinitions := strings.Split(
				strings.TrimPrefix(
					strings.TrimSuffix(definition, ">"),
					"array<",
				),
				"|",
			)

			for _, elementDefinition := range elementDefinitions {
				elementAnyOf = append(elementAnyOf, as.DeclareDataType(elementDefinition))
			}
		}

		return &ArrayDataType{abstractDataType: abstractDataType{definition}, elementAnyOf: elementAnyOf}
	case isScalar(definition):
		return &ScalarDataType{abstractDataType: abstractDataType{definition}}
	default:
		return &ObjectDataType{abstractDataType: abstractDataType{definition}}
	}
}

func isScalar(definition string) bool {
	switch definition {
	case "string", "int32", "int64", "float", "boolean":
		return true
	}

	return false
}
