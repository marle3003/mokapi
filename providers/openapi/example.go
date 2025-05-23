package openapi

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type ExampleValue struct {
	Value any
}

type Examples map[string]*ExampleRef

type ExampleRef struct {
	dynamic.Reference
	Value *Example
}

type Example struct {
	Summary       string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description   string `yaml:"description,omitempty" json:"description,omitempty"`
	Value         any    `yaml:"value,omitempty" json:"value,omitempty"`
	ExternalValue string `yaml:"externalValue,omitempty" json:"externalValue,omitempty"`
}

func (r *ExampleRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ExampleRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (e *ExampleValue) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &e.Value)
}

func (e *Example) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected yaml.MappingNode, got %T", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		v := node.Content[i+1]
		var err error
		switch key {
		case "summary":
			err = v.Decode(&e.Summary)
		case "description":
			err = v.Decode(&e.Description)
		case "value":
			e.Value, err = dynamic.ParseYamlPlain(v)
		case "externalValue":
			err = v.Decode(&e.ExternalValue)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ExampleValue) UnmarshalYAML(node *yaml.Node) error {
	v, err := dynamic.ParseYamlPlain(node)
	if err != nil {
		return err
	}
	e.Value = v
	return nil
}

func (e Examples) parse(config *dynamic.Config, reader dynamic.Reader) error {
	for name, ex := range e {
		if err := ex.parse(config, reader); err != nil {
			return fmt.Errorf("parse example '%v' failed: %w", name, err)
		}
	}

	return nil
}

func (r *ExampleRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	return r.Value.parse(config, reader)
}

func (e *Example) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if e == nil {
		return nil
	}

	if e.ExternalValue != "" {
		return dynamic.Resolve(e.ExternalValue, &e.Value, config, reader)
	}

	return nil
}

func (e Examples) patch(patch Examples) {
	for k, p := range patch {
		if p == nil || p.Value == nil {
			continue
		}
		if v, ok := e[k]; ok && v != nil {
			v.patch(p)
		} else {
			e[k] = p
		}
	}
}

func (r *ExampleRef) patch(patch *ExampleRef) {
	if r.Value == nil {
		r.Value = patch.Value
		return
	}

	if len(patch.Value.Summary) > 0 {
		r.Value.Summary = patch.Value.Summary
	}

	if patch.Value.Value != nil {
		r.Value.Value = patch.Value.Value
	}

	if len(patch.Value.Description) > 0 {
		r.Value.Description = patch.Value.Description
	}

	if len(patch.Value.ExternalValue) > 0 {
		r.Value.ExternalValue = patch.Value.ExternalValue
	}
}
