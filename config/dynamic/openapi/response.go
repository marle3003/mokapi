package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/media"
	"mokapi/sortedmap"
	"strconv"
)

type Responses struct {
	sortedmap.LinkedHashMap[int, *ResponseRef]
} // map[HttpStatus]*ResponseRef

type ResponseRef struct {
	ref.Reference
	Value *Response
}

type Response struct {
	// A short description of the response. CommonMark syntax
	// MAY be used for rich text representation.
	Description string

	// A map containing descriptions of potential response payloads.
	// The key is a media type or media type range and the value describes
	// it. For responses that match multiple keys, only the most specific
	// key is applicable. e.g. text/plain overrides text/*
	Content Content

	// Maps a header name to its definition. RFC7230 states header names are
	// case-insensitive. If a response header is defined with the name
	// "Content-Type", it SHALL be ignored.
	Headers Headers
}

func (r *Responses) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected openapi.Responses map, got %s", token)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[int, *ResponseRef]{}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &ResponseRef{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}

		if key == "default" {
			r.Set(0, val)
		} else {
			statusCode, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("unable to parse http status %v", key)
			}
			r.Set(statusCode, val)
		}
	}
}

func (r *ResponseRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *Responses) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected openapi.Responses map, got %v", value.Tag)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[int, *ResponseRef]{}
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		val := &ResponseRef{}
		err = value.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}

		if key == "default" {
			r.Set(0, val)
		} else {
			statusCode, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("unable to parse http status %v", key)
			}
			r.Set(statusCode, val)
		}
	}

	return nil
}

func (r *ResponseRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *Responses) GetResponse(httpStatus int) *Response {
	res := r.Get(httpStatus)
	if res == nil {
		// 0 as default
		res = r.Get(0)
	}

	if res == nil {
		return nil
	}

	return res.Value
}

func (r *Response) GetContent(contentType media.ContentType) *MediaType {
	for _, v := range r.Content {
		if v.ContentType.Match(contentType) {
			return v
		}
	}

	return nil
}

func (r *Responses) parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	for it := r.Iter(); it.Next(); {
		res := it.Value()
		if err := res.parse(config, reader); err != nil {
			return fmt.Errorf("parse response '%v' failed: %w", it.Key(), err)
		}
	}

	return nil
}

func (r *ResponseRef) parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	return r.Value.parse(config, reader)
}

func (r *Response) parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if err := r.Headers.parse(config, reader); err != nil {
		return err
	}

	return r.Content.parse(config, reader)
}

func (r *Responses) patch(patch *Responses) {
	if patch == nil {
		return
	}

	for it := patch.Iter(); it.Next(); {
		res := it.Value()
		if res.Value == nil {
			continue
		}
		statusCode := it.Key()
		if v := r.GetResponse(statusCode); v != nil {
			v.patch(res.Value)
		} else {
			r.Set(statusCode, res)
		}
	}
}

func (r *Response) patch(patch *Response) {
	if len(patch.Description) > 0 {
		r.Description = patch.Description
	}

	if r.Content == nil {
		r.Content = patch.Content
	} else {
		r.Content.patch(patch.Content)
	}

	if len(r.Headers) == 0 {
		r.Headers = patch.Headers
	} else {
	}
}
