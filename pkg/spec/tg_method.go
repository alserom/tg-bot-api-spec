package spec

import (
	"sync"
)

type TgMethodSpec struct {
	category    string
	name        string
	link        string
	description string
	arguments   []*TgMethodSpecArgument
	returns     []DataTypeDefinition
	a_mu        *sync.RWMutex
	r_mu        *sync.RWMutex
}

func (tms TgMethodSpec) GetCategory() string {
	return tms.category
}

func (tms TgMethodSpec) GetName() string {
	return tms.name
}

func (tms TgMethodSpec) GetLink() string {
	return tms.link
}

func (tms *TgMethodSpec) SetDescription(description string) {
	tms.description = description
}

func (tms TgMethodSpec) GetDescription() string {
	return tms.description
}

func (tms *TgMethodSpec) AddArgument(argument *TgMethodSpecArgument) error {
	if argument == nil {
		return skippedAddingNilPoiner()
	}

	tms.a_mu.Lock()
	tms.arguments = append(tms.arguments, argument)
	tms.a_mu.Unlock()

	return nil
}

func (tms TgMethodSpec) GetArguments() []*TgMethodSpecArgument {
	return tms.arguments
}

func (tms *TgMethodSpec) AddReturnType(returnType DataTypeDefinition) error {
	if returnType == nil {
		return skippedAddingNilPoiner()
	}

	tms.r_mu.Lock()
	tms.returns = append(tms.returns, returnType)
	tms.r_mu.Unlock()

	return nil
}

func (tms TgMethodSpec) GetReturnTypes() []DataTypeDefinition {
	return tms.returns
}

type TgMethodSpecArgument struct {
	name        string
	description string
	required    bool
	dataTypes   []DataTypeDefinition
	dt_mu       *sync.RWMutex
}

func (tmsa TgMethodSpecArgument) GetName() string {
	return tmsa.name
}

func (tmsa *TgMethodSpecArgument) SetDescription(description string) {
	tmsa.description = description
}

func (tmsa TgMethodSpecArgument) GetDescription() string {
	return tmsa.description
}

func (tmsa *TgMethodSpecArgument) SetRequired(required bool) {
	tmsa.required = required
}

func (tmsa TgMethodSpecArgument) IsRequired() bool {
	return tmsa.required
}

func (tmsa *TgMethodSpecArgument) AddDataType(typeDef DataTypeDefinition) error {
	if typeDef == nil {
		return skippedAddingNilPoiner()
	}

	tmsa.dt_mu.Lock()
	tmsa.dataTypes = append(tmsa.dataTypes, typeDef)
	tmsa.dt_mu.Unlock()

	return nil
}

func (tmsa TgMethodSpecArgument) GetDataTypes() []DataTypeDefinition {
	return tmsa.dataTypes
}

func NewTgMethodSpec(category, name, link string) (*TgMethodSpec, error) {
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

	return &TgMethodSpec{
		category: category,
		name:     name,
		link:     link,
		a_mu:     &sync.RWMutex{},
		r_mu:     &sync.RWMutex{},
	}, nil
}

func NewTgMethodSpecArgument(name string) (*TgMethodSpecArgument, error) {
	err := validateNonEmptyStringArg("name", name)
	if err != nil {
		return nil, err
	}

	return &TgMethodSpecArgument{
		name:     name,
		required: true,
		dt_mu:    &sync.RWMutex{},
	}, nil
}
