package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/media"
	"mokapi/sortedmap"
	"strconv"
)

type Responses[K string | int] struct {
	sortedmap.LinkedHashMap[K, *ResponseRef]
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

func (r *Responses[K]) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected openapi.Responses map, got %s", token)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[K, *ResponseRef]{}
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
		switch m := any(&r.LinkedHashMap).(type) {
		case *sortedmap.LinkedHashMap[string, *ResponseRef]:
			m.Set(key, val)
		case *sortedmap.LinkedHashMap[int, *ResponseRef]:
			if key == "default" {
				m.Set(0, val)
			} else {
				statusCode, err := strconv.Atoi(key)
				if err != nil {
					return fmt.Errorf("unable to parse http status %v", key)
				}
				m.Set(statusCode, val)
			}
		}
	}
}

func (r *ResponseRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *Responses[K]) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected openapi.Responses map, got %v", value.Tag)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[K, *ResponseRef]{}
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
		switch m := any(&r.LinkedHashMap).(type) {
		case *sortedmap.LinkedHashMap[string, *ResponseRef]:
			m.Set(key, val)
		case *sortedmap.LinkedHashMap[int, *ResponseRef]:
			if key == "default" {
				m.Set(0, val)
			} else {
				statusCode, err := strconv.Atoi(key)
				if err != nil {
					return fmt.Errorf("unable to parse http status %v", key)
				}
				m.Set(statusCode, val)
			}
		}
	}

	return nil
}

func (r *ResponseRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *Responses[K]) Resolve(token string) (interface{}, error) {
	var res *ResponseRef
	switch m := any(&r.LinkedHashMap).(type) {
	case *sortedmap.LinkedHashMap[int, *ResponseRef]:
		i, err := strconv.Atoi(token)
		if err != nil {
			return nil, err
		}
		res, _ = m.Get(i)
	case *sortedmap.LinkedHashMap[string, *ResponseRef]:
		res, _ = m.Get(token)
	}
	if res == nil {
		return nil, fmt.Errorf("unable to resolve %v", token)
	}
	return res.Value, nil
}

func (r *Responses[K]) GetResponse(httpStatus K) *Response {
	res, _ := r.Get(httpStatus)
	if res == nil {
		switch m := any(&r.LinkedHashMap).(type) {
		case *sortedmap.LinkedHashMap[int, *ResponseRef]:
			res, _ = m.Get(0)
		}
	}

	if res == nil {
		return nil
	}

	return res.Value
}

func (r *Response) GetContent(contentType media.ContentType) *MediaType {
	var found *MediaType

	for _, v := range r.Content {
		if v.ContentType.Match(contentType) {
			found = getBestMediaType(found, v)
		}
	}

	return found
}

func getBestMediaType(m1, m2 *MediaType) *MediaType {
	if m1 == nil {
		return m2
	}
	if !m1.ContentType.IsAny() && !m1.ContentType.IsRange() {
		return m1
	}
	if !m2.ContentType.IsAny() && !m2.ContentType.IsRange() {
		return m2
	}
	if !m1.ContentType.IsAny() {
		return m1
	}
	if !m2.ContentType.IsAny() {
		return m2
	}
	return m1
}

func (r *Responses[K]) parse(config *dynamic.Config, reader dynamic.Reader) error {
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

func (r *ResponseRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	return r.Value.parse(config, reader)
}

func (r *Response) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if err := r.Headers.parse(config, reader); err != nil {
		return err
	}

	return r.Content.parse(config, reader)
}

func (r *Responses[K]) patch(patch *Responses[K]) {
	if patch == nil {
		return
	}

	for it := patch.Iter(); it.Next(); {
		res := it.Value()
		if res.Value == nil {
			continue
		}
		statusCode := it.Key()
		if v, _ := r.Get(statusCode); v != nil && v.Value != nil {
			v.Value.patch(res.Value)
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
		r.Headers.patch(patch.Headers)
	}
}
