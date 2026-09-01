package bruno

import (
	"mokapi/version"
)

type Collection struct {
	Version *version.Version `yaml:"opencollection" json:"opencollection"`
	Info    Info             `yaml:"info,omitempty" json:"info,omitempty"`
	Config  *Config          `yaml:"config,omitempty" json:"config,omitempty"`
	Items   []any            `yaml:"items,omitempty" json:"items,omitempty"`
	Request *RequestDefault  `yaml:"request,omitempty" json:"request,omitempty"`
	Bundled bool             `yaml:"bundled" json:"bundled"`
}

type Info struct {
	Name    string   `yaml:"name" json:"name"`
	Summary string   `yaml:"summary,omitempty" json:"summary,omitempty"`
	Version string   `yaml:"version,omitempty" json:"version,omitempty"`
	Authors []Author `yaml:"authors,omitempty" json:"authors,omitempty"`
}

type Author struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
	Url   string `yaml:"url" json:"url"`
}

type Config struct {
	Environments []Environment `yaml:"environments,omitempty" json:"environments,omitempty"`
}

type Environment struct {
	Name        string     `yaml:"name" json:"name"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   []Variable `yaml:"variables,omitempty" json:"variables,omitempty"`
}

type Variable struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

type FolderItem struct {
	Info  *FolderInfo `yaml:"info,omitempty" json:"info,omitempty"`
	Items []any       `yaml:"items,omitempty" json:"items,omitempty"`
}

type FolderInfo struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string `yaml:"type" json:"type"`
	Sequence    int    `yaml:"seq,omitempty" json:"seq,omitempty"`
}

type HttpItem struct {
	Info *HttpInfo   `yaml:"info,omitempty" json:"info,omitempty"`
	Http *HttpDetail `yaml:"http,omitempty" json:"http,omitempty"`
}

type HttpInfo struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string   `yaml:"type,omitempty" json:"type,omitempty"`
	Sequence    int      `yaml:"seq,omitempty" json:"seq,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

type HttpDetail struct {
	Method  string              `yaml:"method" json:"method"`
	Url     string              `yaml:"url" json:"url"`
	Headers []HttpRequestHeader `yaml:"headers,omitempty" json:"headers,omitempty"`
	Params  []HttpRequestParam  `yaml:"params,omitempty" json:"params,omitempty"`
	Body    *HttpRequestBody    `yaml:"body,omitempty" json:"body,omitempty"`
}

type HttpRequestHeader struct {
	Name        string `yaml:"name" json:"name"`
	Value       string `yaml:"value" json:"value"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Disabled    bool   `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

type HttpRequestParam struct {
	Name        string `yaml:"name" json:"name"`
	Value       string `yaml:"value" json:"value"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string `yaml:"type" json:"type"`
	Disabled    bool   `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

type HttpRequestBody struct {
	Body    *HttpRequestBodyRaw
	Variant []HttpRequestBodyVariant
}

type HttpRequestBodyRaw struct {
	Type string `yaml:"type" json:"type"`
	Data string `yaml:"data" json:"data"`
}

type HttpRequestBodyVariant struct {
	Title    string             `yaml:"title" json:"title"`
	Selected bool               `yaml:"selected" json:"selected"`
	Body     HttpRequestBodyRaw `yaml:"body" json:"body"`
}

type RequestDefault struct {
	Variables []Variable `yaml:"variables,omitempty" json:"variables,omitempty"`
}

func (b *HttpRequestBody) MarshalYAML() (interface{}, error) {
	if b.Body != nil {
		return b.Body, nil
	}
	return b.Variant, nil
}
