package openapi

type Encoding struct {
	ContentType   string            `yaml:"contentType,omitempty" json:"contentType,omitempty"`
	Headers       map[string]Header `yaml:"headers,omitempty" json:"headers,omitempty"`
	Style         string            `yaml:"style,omitempty" json:"style,omitempty"`
	Explode       bool              `yaml:"explode,omitempty" json:"explode,omitempty"`
	AllowReserved bool              `yaml:"allowReserved,omitempty" json:"allowReserved,omitempty"`
}
