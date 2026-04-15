package dynamic

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Reference[T any] struct {
	Ref        string `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	DynamicRef string `yaml:"$dynamicRef,omitempty" json:"$dynamicRef,omitempty"`

	Summary     string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	origin *Config
}

func (r *Reference[T]) UnmarshalYaml(node *yaml.Node, val interface{}) error {
	err := node.Decode(r)
	if err == nil && len(r.Ref) > 0 {
		return nil
	}

	return node.Decode(val)
}

func (r *Reference[T]) UnmarshalJson(b []byte, val interface{}) error {
	var m map[string]string
	_ = json.Unmarshal(b, &m)
	if _, ok := m["$ref"]; ok {
		return UnmarshalJSON(b, r)
	}

	err := UnmarshalJSON(b, val)
	return err
}

func (r *Reference[T]) Parse(config *Config, _ Reader) error {
	if r.Ref == "" || r.origin != nil {
		return nil
	}
	r.origin = config
	return nil
}

func (r *Reference[T]) HasRef() bool {
	return r.Ref != "" || r.DynamicRef != ""
}

func (r *Reference[T]) Resolve(config *Config, reader Reader) (T, error) {
	var err error
	var result T

	if err := r.Parse(config, reader); err != nil {
		return result, err
	}

	if r.Ref != "" {
		ref := r.Ref
		if !strings.HasPrefix(ref, "#") {
			u, err := resolveUrl(r.Ref, r.origin)
			if err != nil {
				return result, fmt.Errorf("resolve reference '%s' failed: %v", r.Ref, err)
			}
			ref = u.String()
		}

		result, err = resolve[T](ref, config, reader)
		return result, err
	}

	result, err = ResolveDynamic[T](r.DynamicRef, config, reader)
	return result, err
}
