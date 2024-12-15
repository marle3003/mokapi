package openapi

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/providers/openapi/parameter"
	"net/http"
)

var NoSuccessResponse = errors.New("no success response (HTTP 2xx) in configuration")

type Operation struct {
	// A list of tags for API documentation control. Tags can be used for
	// logical grouping of operations by resources or any other qualifier.
	Tags []string `yaml:"tags" json:"tags"`

	// A short summary of what the operation does.
	Summary string `yaml:"summary" json:"summary"`

	// A verbose explanation of the operation behavior.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `yaml:"description" json:"description"`

	Deprecated bool `yaml:"deprecated" json:"deprecated"`

	// Unique string used to identify the operation. The id MUST be unique
	// among all operations described in the API. The operationId value is
	// case-sensitive. Tools and libraries MAY use the operationId to uniquely
	// identify an operation, therefore, it is RECOMMENDED to follow common
	// programming naming conventions.
	OperationId string `yaml:"operationId" json:"operationId"`

	// A list of parameters that are applicable for this operation.
	// If a parameter is already defined at the Path Item, the new definition
	// will override it but can never remove it. The list MUST NOT include
	// duplicated parameters. A unique parameter is defined by a combination
	// of a name and location
	Parameters parameter.Parameters

	RequestBody *RequestBodyRef `yaml:"requestBody" json:"requestBody"`

	// The list of possible responses as they are returned from executing this
	// operation.
	Responses *Responses[int] `yaml:"responses" json:"responses"`

	Path *Path `yaml:"-" json:"-"`
}

func (o *Operation) getFirstSuccessResponse() (int, *Response, error) {
	for it := o.Responses.Iter(); it.Next(); {
		status := it.Key()
		if IsHttpStatusSuccess(status) {
			r := it.Value()
			if r != nil {
				return status, r.Value, nil
			}
			return status, nil, nil
		}
	}

	return 0, nil, NoSuccessResponse
}

func getDefaultResponse() (int, *Response) {
	r := &Response{
		Content: Content{
			"application/json": &MediaType{
				Schema:      nil,
				ContentType: media.ParseContentType("application/json"),
			},
		},
	}

	return http.StatusOK, r
}

func (o *Operation) getResponse(statusCode int) *Response {
	return o.Responses.GetResponse(statusCode)
}

func (o *Operation) parse(p *Path, config *dynamic.Config, reader dynamic.Reader) error {
	if o == nil {
		return nil
	}

	o.Path = p

	if err := o.Parameters.Parse(config, reader); err != nil {
		return err
	}

	if err := o.RequestBody.parse(config, reader); err != nil {
		return fmt.Errorf("parse request body failed: %w", err)
	}

	return o.Responses.parse(config, reader)
}

func (o *Operation) patch(patch *Operation) {
	if patch == nil {
		return
	}

	if len(patch.Summary) > 0 {
		o.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		o.Description = patch.Description
	}
	if len(patch.OperationId) > 0 {
		o.OperationId = patch.OperationId
	}
	o.Deprecated = patch.Deprecated

	if o.RequestBody == nil {
		o.RequestBody = patch.RequestBody
	} else {
		o.RequestBody.patch(patch.RequestBody)
	}

	if o.Responses == nil {
		o.Responses = patch.Responses
	} else {
		o.Responses.patch(patch.Responses)
	}
}
