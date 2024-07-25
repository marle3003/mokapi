package openapi

import (
	_ "embed"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/feature"
	"mokapi/media"
	"mokapi/schema/json/parser"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/version"
	"net/http"
)

var (
	supportedVersions = []version.Version{
		version.New("3.0.0"),
		version.New("3.0.1"),
		version.New("3.0.2"),
		version.New("3.0.3"),
		version.New("3.1.0"),
	}

	//go:embed schema.yaml
	validation_schema_raw []byte
	validation_schema     *jsonSchema.Ref
)

func init() {
	err := yaml.Unmarshal(validation_schema_raw, &validation_schema)
	if err != nil {
		panic(err)
	}
	err = validation_schema.Parse(&dynamic.Config{Data: validation_schema}, nil)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	OpenApi version.Version `yaml:"openapi" json:"openapi"`
	Info    Info            `yaml:"info" json:"info"`
	Servers []*Server       `yaml:"servers,omitempty" json:"servers,omitempty"`

	// A relative path to an individual endpoint. The path MUST begin
	// with a forward slash ('/'). The path is appended to the url from
	// server objects url field in order to construct the full URL
	Paths      PathItems  `yaml:"paths,omitempty" json:"paths,omitempty"`
	Components Components `yaml:"components,omitempty" json:"components,omitempty"`

	ExternalDocs *ExternalDocs `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

type ExternalDocs struct {
	Description string `yaml:"description" json:"description"`
	Url         string `yaml:"url" json:"url"`
}

func IsHttpStatusSuccess(status int) bool {
	return status == http.StatusOK ||
		status == http.StatusCreated ||
		status == http.StatusAccepted ||
		status == http.StatusNonAuthoritativeInfo ||
		status == http.StatusNoContent ||
		status == http.StatusResetContent ||
		status == http.StatusPartialContent ||
		status == http.StatusMultiStatus ||
		status == http.StatusAlreadyReported ||
		status == http.StatusIMUsed
}

func (c *Config) Validate() (err error) {
	if c.OpenApi.IsEmpty() {
		err = errors.Join(err, fmt.Errorf("no OpenApi version defined"))
	} else {
		if !c.OpenApi.IsSupported(supportedVersions...) {
			err = errors.Join(err, fmt.Errorf("not supported version: %v", &c.OpenApi))
		}
	}

	if len(c.Info.Name) == 0 {
		err = errors.Join(err, errors.New("an openapi title is required"))
	}

	if feature.IsEnabled("openapi-validation") {
		p := &parser.Parser{}
		_, errParse := p.Parse(c, validation_schema)
		if errParse != nil {
			err = errors.Join(err, errParse)
		}
	}

	return err
}

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if c == nil {
		return nil
	}

	if err := c.Components.parse(config, reader); err != nil {
		return err
	}

	if err := c.Paths.parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (c *Config) Patch(patch *Config) {
	c.Info.patch(patch.Info)
	c.patchServers(patch.Servers)
	if c.Paths == nil {
		c.Paths = patch.Paths
	} else {
		c.Paths.patch(patch.Paths)
	}
	c.Components.patch(patch.Components)
}

func (r *RequestBody) GetMedia(contentType media.ContentType) *MediaType {
	for _, v := range r.Content {
		if v.ContentType.Match(contentType) {
			return v
		}
	}

	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	type alias Config
	a := alias(*c)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*c = Config(a)
	return nil
}
