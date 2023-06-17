package export_to_json

type JsonData struct {
	Version     string              `json:"version"`
	ReleaseDate string              `json:"releaseDate"`
	Link        string              `json:"link"`
	Types       map[string]TgType   `json:"types"`
	Methods     map[string]TgMethod `json:"methods"`
}

type TgType struct {
	Category    string           `json:"category"`
	Name        string           `json:"name"`
	Link        string           `json:"link"`
	Description string           `json:"description"`
	Parent      *NilableString   `json:"parent,omitempty"`
	Children    []string         `json:"children,omitempty"`
	Properties  []TgTypeProperty `json:"properties,omitempty"`
}

type TgTypeProperty struct {
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Types           []string       `json:"types"`
	Optional        bool           `json:"optional"`
	PredefinedValue *NilableString `json:"default,omitempty"`
}

type TgMethod struct {
	Category    string             `json:"category"`
	Name        string             `json:"name"`
	Link        string             `json:"link"`
	Description string             `json:"description"`
	Arguments   []TgMethodArgument `json:"arguments,omitempty"`
	Returns     []string           `json:"returns"`
}

type TgMethodArgument struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Types       []string `json:"types"`
}

type NilableString string
