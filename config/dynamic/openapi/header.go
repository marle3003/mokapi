package openapi

import (
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

func (r *HeaderRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
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
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
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
