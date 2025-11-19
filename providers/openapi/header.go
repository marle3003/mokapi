package openapi

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/js/util"
	"mokapi/providers/openapi/schema"
	"net/http"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Headers map[string]*HeaderRef

type HeaderRef struct {
	dynamic.Reference
	Value *Header
}

type Header struct {
	Parameter
}

func (r *HeaderRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (h *Header) UnmarshalJSON(b []byte) error {
	type alias Header
	header := alias{}
	err := json.Unmarshal(b, &header)
	if err != nil {
		return err
	}
	header.Type = ParameterHeader
	*h = Header(header)
	return nil
}

func (r *HeaderRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (h *Header) UnmarshalYAML(node *yaml.Node) error {
	type alias Header
	header := alias{}
	err := node.Decode(&header)
	if err != nil {
		return err
	}
	header.Type = ParameterHeader
	*h = Header(header)
	return nil
}

func (h Headers) parse(config *dynamic.Config, reader dynamic.Reader) error {
	for name, header := range h {
		if err := header.parse(config, reader); err != nil {
			return fmt.Errorf("parse header '%v' failed: %w", name, err)
		}
	}

	return nil
}

func (r *HeaderRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}
	return r.Value.Parse(config, reader)
}

func (h Headers) patch(patch Headers) {
	for k, p := range patch {
		if p == nil || p.Value == nil {
			continue
		}
		if v, ok := h[k]; ok && v != nil {
			v.patch(p)
		} else {
			h[k] = p
		}
	}
}

func (r *HeaderRef) patch(patch *HeaderRef) {
	if patch == nil || patch.Value == nil {
		return
	}

	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.patch(patch.Value)
	}
}

func (h *Header) patch(patch *Header) {
	if len(patch.Name) > 0 {
		h.Name = patch.Name
	}
	if len(patch.Description) > 0 {
		h.Description = patch.Description
	}
	if h.Schema == nil {
		h.Schema = patch.Schema
	} else {
		h.Schema.Patch(patch.Schema)
	}
}

func (h *Header) marshal(v any, rw http.ResponseWriter) error {
	i, err := p.ParseWith(v, schema.ConvertToJsonSchema(h.Schema))
	if err != nil {
		return err
	}
	switch vv := i.(type) {
	case string:
		rw.Header().Set(h.Name, vv)
	case []interface{}:

		for _, item := range vv {
			rw.Header().Add(h.Name, fmt.Sprintf("%v", item))
		}
	default:
		rw.Header().Set(h.Name, fmt.Sprintf("%v", vv))
	}
	return nil
}

func getContentType(headers map[string]any) (string, error) {
	v, err := getHeaderValue(headers["Content-Type"])
	if err != nil {
		return "", fmt.Errorf("invalid header 'Content-Type': %w", err)
	}
	if len(v) > 0 {
		return v[0], nil
	}
	return "", nil
}

func setHeaders(headers map[string]any, definition Headers, rw http.ResponseWriter) error {
	for name, value := range headers {
		switch name {
		case "Content-Type":
			v, err := getHeaderValue(value)
			if err != nil {
				return fmt.Errorf("invalid header '%s': %w", name, err)
			}
			if len(v) > 0 {
				rw.Header().Add(name, v[0])
			}
			continue
		}

		if header, ok := definition[name]; ok && header.Value != nil {
			err := header.Value.marshal(value, rw)
			if err != nil {
				return fmt.Errorf("invalid header '%s': %w", name, err)
			}
		} else {
			v, err := getHeaderValue(value)
			if err != nil {
				return fmt.Errorf("invalid header '%s': %w", name, err)
			}
			for _, item := range v {
				rw.Header().Add(name, item)
			}
		}
	}
	return nil
}

func getHeaderValue(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return []string{v.String()}, nil
	case reflect.Slice:
		var result []string
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}
			result = append(result, item.String())
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected a string or array of strings, but received %s", util.JsType(v.Interface()))
	}
}
