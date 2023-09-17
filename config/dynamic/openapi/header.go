package openapi

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/ref"
)

type Headers map[string]*HeaderRef

type HeaderRef struct {
	ref.Reference
	Value *Header
}

type Header struct {
	parameter.Parameter
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
	header.Type = parameter.Header
	*h = Header(header)
	return nil
}

func (r *HeaderRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (h *Header) UnmarshalYAML(node *yaml.Node) error {
	type alias Header
	header := alias{}
	err := node.Decode(&header)
	if err != nil {
		return err
	}
	header.Type = parameter.Header
	*h = Header(header)
	return nil
}

func (h Headers) parse(config *common.Config, reader common.Reader) error {
	for name, header := range h {
		if err := header.parse(config, reader); err != nil {
			return fmt.Errorf("parse header '%v' failed: %w", name, err)
		}
	}

	return nil
}

func (r *HeaderRef) parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
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
