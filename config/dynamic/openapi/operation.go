package openapi

import (
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/parameter"
)

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
	Responses *Responses `yaml:"responses" json:"responses"`

	Path *Path `yaml:"-" json:"-"`
}

func (o *Operation) getFirstSuccessResponse() (int, *Response, error) {
	var successStatus int
	for it := o.Responses.Iter(); it.Next(); {
		status := it.Key()
		if IsHttpStatusSuccess(status) {
			successStatus = status
			break
		}
	}

	if successStatus == 0 {
		return 0, nil, fmt.Errorf("no success response (HTTP 2xx) in configuration")
	}

	r := o.Responses.GetResponse(successStatus)
	return successStatus, r, nil
}

func (o *Operation) getResponse(statusCode int) *Response {
	return o.Responses.GetResponse(statusCode)
}

func (o *Operation) parse(p *Path, config *common.Config, reader common.Reader) error {
	if o == nil {
		return nil
	}

	o.Path = p

	if err := o.Parameters.Parse(config, reader); err != nil {
		return err
	}

	if err := o.RequestBody.Parse(config, reader); err != nil {
		return err
	}

	return o.Responses.parse(config, reader)
}

func (o *Operation) patch(patch *Operation) {
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
