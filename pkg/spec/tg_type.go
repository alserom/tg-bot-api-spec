package spec

import (
	"sync"
)

type TgTypeSpec struct {
	category    string
	name        string
	link        string
	description string
	parent      *TgTypeSpec
	children    []*TgTypeSpec
	properties  []*TgTypeSpecProperty
	c_mu        *sync.RWMutex
	p_mu        *sync.RWMutex
}

func (tts TgTypeSpec) GetCategory() string {
	return tts.category
}

func (tts TgTypeSpec) GetName() string {
	return tts.name
}

func (tts TgTypeSpec) GetLink() string {
	return tts.link
}

func (tts *TgTypeSpec) SetDescription(description string) {
	tts.description = description
}

func (tts TgTypeSpec) GetDescription() string {
	return tts.description
}

func (tts *TgTypeSpec) SetParent(parent *TgTypeSpec) {
	tts.parent = parent
}

func (tts TgTypeSpec) GetParent() *TgTypeSpec {
	return tts.parent
}

func (tts *TgTypeSpec) AddChild(child *TgTypeSpec) error {
	if child == nil {
		return skippedAddingNilPoiner()
	}

	tts.c_mu.Lock()
	tts.children = append(tts.children, child)
	tts.c_mu.Unlock()

	return nil
}

func (tts TgTypeSpec) GetChildren() []*TgTypeSpec {
	return tts.children
}

func (tts *TgTypeSpec) AddProperty(property *TgTypeSpecProperty) error {
	if property == nil {
		return skippedAddingNilPoiner()
	}

	tts.p_mu.Lock()
	tts.properties = append(tts.properties, property)
	tts.p_mu.Unlock()

	return nil
}

func (tts TgTypeSpec) GetProperties() []*TgTypeSpecProperty {
	return tts.properties
}

type TgTypeSpecProperty struct {
	name            string
	description     string
	dataTypes       []DataTypeDefinition
	optional        bool
	predefinedValue *TgTypeSpecPropertyValue
	dt_mu           *sync.RWMutex
}

func (ttsp TgTypeSpecProperty) GetName() string {
	return ttsp.name
}

func (ttsp *TgTypeSpecProperty) SetDescription(description string) {
	ttsp.description = description
}

func (ttsp TgTypeSpecProperty) GetDescription() string {
	return ttsp.description
}

func (ttsp *TgTypeSpecProperty) AddDataType(typeDefinition DataTypeDefinition) error {
	if typeDefinition == nil {
		return skippedAddingNilPoiner()
	}

	ttsp.dt_mu.Lock()
	ttsp.dataTypes = append(ttsp.dataTypes, typeDefinition)
	ttsp.dt_mu.Unlock()

	return nil
}

func (ttsp TgTypeSpecProperty) GetDataTypes() []DataTypeDefinition {
	return ttsp.dataTypes
}

func (ttsp *TgTypeSpecProperty) SetOptional(optional bool) {
	ttsp.optional = optional
}

func (ttsp TgTypeSpecProperty) IsOptional() bool {
	return ttsp.optional
}

func (ttsp *TgTypeSpecProperty) SetPredefinedValue(value *TgTypeSpecPropertyValue) {
	ttsp.predefinedValue = value
}

func (ttsp TgTypeSpecProperty) GetPredefinedValue() *TgTypeSpecPropertyValue {
	return ttsp.predefinedValue
}

type TgTypeSpecPropertyValue string

func NewTgTypeSpec(category, name, link string) (*TgTypeSpec, error) {
	var errs []error
	checks := [3]error{
		validateNonEmptyStringArg("category", category),
		validateNonEmptyStringArg("name", name),
		validateLinkArg("link", link),
	}
	for _, err := range checks {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return nil, &CompositeError{errs}
	}

	return &TgTypeSpec{
		category: category,
		name:     name,
		link:     link,
		c_mu:     &sync.RWMutex{},
		p_mu:     &sync.RWMutex{},
	}, nil
}

func NewTgTypeSpecProperty(name string) (*TgTypeSpecProperty, error) {
	err := validateNonEmptyStringArg("name", name)
	if err != nil {
		return nil, err
	}

	return &TgTypeSpecProperty{name: name, dt_mu: &sync.RWMutex{}}, nil
}
