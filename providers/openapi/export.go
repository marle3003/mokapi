package openapi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/schema"
)

type export struct {
	result *Config
}

func (c *Config) Export() *Config {
	e := &export{
		result: &Config{
			OpenApi: c.OpenApi,
			Info:    c.Info,
			Servers: c.Servers,
		},
	}

	e.buildPaths(c)

	return e.result
}

func (e *export) buildPaths(origin *Config) {
	if origin.Paths == nil || origin.Paths.Len() == 0 {
		return
	}

	e.result.Paths = &PathItems{}
	for it := origin.Paths.Iter(); it.Next(); {
		s := it.Value()
		if s.Value == nil {
			continue
		}
		e.result.Paths.Set(it.Key(), e.buildPath(s))
	}
}

func (e *export) buildPath(origin *PathRef) *PathRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *PathRef
	var value Path
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		if e.result.Components.PathItems == nil {
			e.result.Components.PathItems = &PathItems{}
		}

		ref := fmt.Sprintf("#/components/pathItems/%s", name)
		result = &PathRef{
			Reference: dynamic.Reference[*PathRef]{Ref: ref},
		}

		if _, ok := e.result.Components.PathItems.Get(name); !ok {
			value = *origin.Value
			e.result.Components.PathItems.Set(name, &PathRef{Value: &value})
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &PathRef{Value: &value}
	}

	value.Get = e.buildOperation(origin.Value.Get)
	value.Post = e.buildOperation(origin.Value.Post)
	value.Put = e.buildOperation(origin.Value.Put)
	value.Patch = e.buildOperation(origin.Value.Patch)
	value.Delete = e.buildOperation(origin.Value.Delete)
	value.Head = e.buildOperation(origin.Value.Head)
	value.Options = e.buildOperation(origin.Value.Options)
	value.Trace = e.buildOperation(origin.Value.Trace)
	value.Query = e.buildOperation(origin.Value.Query)

	value.AdditionalOperations = map[string]*Operation{}
	for k, v := range origin.Value.AdditionalOperations {
		value.AdditionalOperations[k] = e.buildOperation(v)
	}

	value.Parameters = nil
	for _, param := range origin.Value.Parameters {
		value.Parameters = append(value.Parameters, e.buildParameter(param))
	}

	return result
}

func (e *export) buildOperation(origin *Operation) *Operation {
	if origin == nil {
		return nil
	}

	op := *origin

	op.Parameters = nil
	for _, param := range origin.Parameters {
		op.Parameters = append(op.Parameters, e.buildParameter(param))
	}

	op.RequestBody = e.buildRequestBody(origin.RequestBody)

	if origin.Responses != nil {
		op.Responses = &Responses{}
		for it := origin.Responses.Iter(); it.Next(); {
			op.Responses.Set(it.Key(), e.buildResponse(it.Value()))
		}
	}

	return &op
}

func (e *export) buildParameter(origin *ParameterRef) *ParameterRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *ParameterRef
	var value Parameter
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		if e.result.Components.Parameters == nil {
			e.result.Components.Parameters = ComponentParameters{}
		}

		ref := fmt.Sprintf("#/components/parameters/%s", name)
		result = &ParameterRef{
			Reference: dynamic.Reference[*ParameterRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Parameters[name]; !ok {
			value = *origin.Value
			e.result.Components.Parameters[name] = &ParameterRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &ParameterRef{Value: &value}
	}

	value.Schema = e.buildSchema(origin.Value.Schema)

	value.Content = Content{}
	for k, v := range origin.Value.Content {
		value.Content[k] = e.buildContent(v)
	}

	return result
}

func (e *export) buildRequestBody(origin *RequestBodyRef) *RequestBodyRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *RequestBodyRef
	var value RequestBody
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		if e.result.Components.RequestBodies == nil {
			e.result.Components.RequestBodies = RequestBodies{}
		}

		ref := fmt.Sprintf("#/components/requestBodies/%s", name)
		result = &RequestBodyRef{
			Reference: dynamic.Reference[*RequestBodyRef]{Ref: ref},
		}

		if _, ok := e.result.Components.RequestBodies[name]; !ok {
			value = *origin.Value
			e.result.Components.RequestBodies[name] = &RequestBodyRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &RequestBodyRef{Value: &value}
	}

	value.Content = Content{}
	for k, v := range origin.Value.Content {
		value.Content[k] = e.buildContent(v)
	}

	return result
}

func (e *export) buildResponse(origin *ResponseRef) *ResponseRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *ResponseRef
	var value Response
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		if e.result.Components.Responses == nil {
			e.result.Components.Responses = ResponseBodies{}
		}

		ref := fmt.Sprintf("#/components/responses/%s", name)
		result = &ResponseRef{
			Reference: dynamic.Reference[*ResponseRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Responses[name]; !ok {
			value = *origin.Value
			e.result.Components.Responses[name] = &ResponseRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &ResponseRef{Value: &value}
	}

	value.Content = Content{}
	for k, v := range origin.Value.Content {
		value.Content[k] = e.buildContent(v)
	}

	return result
}

func (e *export) buildContent(origin *MediaType) *MediaType {
	if origin == nil {
		return nil
	}

	value := *origin
	value.Schema = e.buildSchema(value.Schema)

	return &value
}

func (e *export) buildSchema(origin *schema.Schema) *schema.Schema {
	if origin == nil {
		return nil
	}

	var result *schema.Schema
	var value schema.Schema
	if origin.Ref != "" {
		name := dynamic.RefName(origin.Ref)
		if e.result.Components.Schemas == nil {
			e.result.Components.Schemas = &schema.Schemas{}
		}

		ref := fmt.Sprintf("#/components/schemas/%s", name)
		result = &schema.Schema{
			Reference: dynamic.Reference[*schema.Schema]{Ref: ref},
		}

		if e.result.Components.Schemas.Get(name) == nil {
			value = *origin
			value.Ref = ""
			e.result.Components.Schemas.Set(name, &value)
		} else {
			return result
		}
	} else {
		value = *origin
		result = &value
	}

	return result
}
