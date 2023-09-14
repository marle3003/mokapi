package openapi

import (
	"errors"
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	common.Register("openapi", &Config{})
}

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

type Content map[string]*MediaType

type MediaType struct {
	Schema   *schema.Ref
	Example  interface{}
	Examples map[string]*ExampleRef

	ContentType media.ContentType `yaml:"-" json:"-"`
}

type HeaderRef struct {
	ref.Reference
	Value *Header
}

type Header struct {
	Name        string
	Description string
	Schema      *schema.Ref
}

type Example struct {
	Summary     string
	Value       interface{}
	Description string
}

type ExampleRef struct {
	ref.Reference
	Value *Example
}

type Components struct {
	Schemas       *schema.SchemasRef         `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Responses     *NamedResponses            `yaml:"responses,omitempty" json:"responses,omitempty"`
	RequestBodies *RequestBodies             `yaml:"requestBodies,omitempty" json:"requestBodies,omitempty"`
	Parameters    *parameter.NamedParameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Examples      *Examples                  `yaml:"examples,omitempty" json:"examples,omitempty"`
	Headers       *NamedHeaders              `yaml:"headers,omitempty" json:"headers,omitempty"`
}

type NamedResponses struct {
	ref.Reference
	Value map[string]*ResponseRef
}

type NamedHeaders struct {
	ref.Reference
	Value map[string]*HeaderRef
}

type Examples struct {
	ref.Reference
	Value map[string]*ExampleRef
}

type RequestBodies struct {
	ref.Reference
	Value map[string]*RequestBodyRef
}

func (c *Config) Validate() error {
	if len(c.OpenApi) == 0 {
		return fmt.Errorf("no OpenApi version defined")
	}
	v := parseVersion(c.OpenApi)
	if v.major != 3 {
		return fmt.Errorf("unsupported version: %v", c.OpenApi)
	}

	if len(c.Info.Name) == 0 {
		return errors.New("an openapi title is required")
	}

	return nil
}

func (r *RequestBody) GetMedia(contentType media.ContentType) *MediaType {
	for _, v := range r.Content {
		if v.ContentType.Match(contentType) {
			return v
		}
	}

	return nil
}

func (r *Response) GetContent(contentType media.ContentType) *MediaType {
	for _, v := range r.Content {
		if v.ContentType.Match(contentType) {
			return v
		}
	}

	return nil
}

type version struct {
	major int
	minor int
	build int
}

func parseVersion(s string) (v version) {
	numbers := strings.Split(s, ".")
	if len(numbers) == 0 {
		return
	}
	if len(numbers) > 0 {
		i, err := strconv.Atoi(numbers[0])
		if err != nil {
			return
		}
		v.major = i
	}
	if len(numbers) > 1 {
		i, err := strconv.Atoi(numbers[1])
		if err != nil {
			return
		}
		v.minor = i
	}
	if len(numbers) > 2 {
		i, err := strconv.Atoi(numbers[2])
		if err != nil {
			return
		}
		v.build = i
	}
	return
}
