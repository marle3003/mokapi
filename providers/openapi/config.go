package openapi

import (
	"errors"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/version"
	"net/http"
)

type Config struct {
	OpenApi string    `yaml:"openapi" json:"openapi"`
	Info    Info      `yaml:"info" json:"info"`
	Servers []*Server `yaml:"servers,omitempty" json:"servers,omitempty"`

	// A relative path to an individual endpoint. The path MUST begin
	// with a forward slash ('/'). The path is appended to the url from
	// server objects url field in order to construct the full URL
	Paths      Paths      `yaml:"paths,omitempty" json:"paths,omitempty"`
	Components Components `yaml:"components,omitempty" json:"components,omitempty"`
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
	if len(c.OpenApi) == 0 {
		return fmt.Errorf("no OpenApi version defined")
	}
	v := version.New(c.OpenApi)
	if v.Major != 3 {
		return fmt.Errorf("unsupported version: %v", c.OpenApi)
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
