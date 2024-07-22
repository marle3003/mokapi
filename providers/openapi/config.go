package openapi

import (
	"errors"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/version"
	"net/http"
)

var supportedVersions = []version.Version{
	version.New("3.0.0"),
	version.New("3.0.1"),
	version.New("3.0.2"),
	version.New("3.0.3"),
	version.New("3.1.0"),
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

func (c *Config) Validate() error {
	if c.OpenApi.IsEmpty() {
		return fmt.Errorf("no OpenApi version defined")
	}
	if !c.OpenApi.IsSupported(supportedVersions...) {
		return fmt.Errorf("not supported version: %v", &c.OpenApi)
	}

	if len(c.Info.Name) == 0 {
		return errors.New("an openapi title is required")
	}

	return nil
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
