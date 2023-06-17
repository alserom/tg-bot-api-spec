package spec

import (
	"sync"
)

type DataSource interface {
	FillApiSpec(as *ApiSpec) error
}

type ApiSpec struct {
	version             string
	releaseDate         string
	link                string
	types               map[string]*TgTypeSpec
	methods             map[string]*TgMethodSpec
	dataTypeDefinitions map[string]DataTypeDefinition
	t_mu                *sync.RWMutex
	m_mu                *sync.RWMutex
	dtd_mu              *sync.RWMutex
}

func (as *ApiSpec) SetVersion(version string) error {
	err := validateNonEmptyStringArg("version", version)
	if err == nil {
		as.version = version
	}

	return err
}

func (as ApiSpec) GetVersion() string {
	return as.version
}

func (as *ApiSpec) SetReleaseDate(date string) error {
	err := validateNonEmptyStringArg("date", date)
	if err == nil {
		as.releaseDate = date
	}

	return err
}

func (as ApiSpec) GetReleaseDate() string {
	return as.releaseDate
}

func (as *ApiSpec) SetLink(link string) error {
	err := validateLinkArg("link", link)
	if err == nil {
		as.link = link
	}

	return err
}

func (as ApiSpec) GetLink() string {
	return as.link
}

func (as *ApiSpec) AddType(t *TgTypeSpec) error {
	if t == nil {
		return skippedAddingNilPoiner()
	}

	as.t_mu.Lock()
	defer as.t_mu.Unlock()

	as.types[t.name] = t

	typeDef := as.DeclareDataType(t.name)
	switch objDef := typeDef.(type) {
	case *ObjectDataType:
		objDef.setRef(t)
	}

	return nil
}

func (as ApiSpec) GetType(name string) (*TgTypeSpec, bool) {
	as.t_mu.RLock()
	item, exists := as.types[name]
	as.t_mu.RUnlock()

	return item, exists
}

func (as ApiSpec) GetTypes() map[string]*TgTypeSpec {
	return as.types
}

func (as *ApiSpec) AddMethod(m *TgMethodSpec) error {
	if m == nil {
		return skippedAddingNilPoiner()
	}

	as.m_mu.Lock()
	as.methods[m.name] = m
	as.m_mu.Unlock()

	return nil
}

func (as ApiSpec) GetMethod(name string) (*TgMethodSpec, bool) {
	as.m_mu.RLock()
	item, exists := as.methods[name]
	as.m_mu.RUnlock()

	return item, exists
}

func (as ApiSpec) GetMethods() map[string]*TgMethodSpec {
	return as.methods
}

func (as *ApiSpec) DeclareDataType(definition string) DataTypeDefinition {
	as.dtd_mu.Lock()
	dataType, exists := as.dataTypeDefinitions[definition]
	as.dtd_mu.Unlock()
	if exists {
		return dataType
	}

	newDataType := newDataTypeDefinition(definition, as)
	as.dtd_mu.Lock()
	as.dataTypeDefinitions[newDataType.GetDefinition()] = newDataType
	as.dtd_mu.Unlock()

	return newDataType
}

func (as ApiSpec) GetDataTypeDefinitions() map[string]DataTypeDefinition {
	return as.dataTypeDefinitions
}

func (as ApiSpec) SelfCheck() error {
	return check(as)
}

func NewApiSpec(ds DataSource) (*ApiSpec, error) {
	as := &ApiSpec{
		types:               make(map[string]*TgTypeSpec),
		methods:             make(map[string]*TgMethodSpec),
		dataTypeDefinitions: make(map[string]DataTypeDefinition),
		t_mu:                &sync.RWMutex{},
		m_mu:                &sync.RWMutex{},
		dtd_mu:              &sync.RWMutex{},
	}

	if err := ds.FillApiSpec(as); err != nil {
		return nil, err
	}

	return as, nil
}
