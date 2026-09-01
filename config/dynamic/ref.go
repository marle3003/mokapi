package dynamic

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
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

func (r *Reference[T]) IsLocalRef() bool {
	return strings.HasPrefix(r.Ref, "#")
}

func (r *Reference[T]) Resolve(config *Config, reader Reader) (T, error) {
	var err error
	var result T

	if err := r.Parse(config, reader); err != nil {
		return result, err
	}

	if r.Ref != "" {
		ref := r.Ref
		if !r.IsLocalRef() {
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

func RefName(ref string) string {
	if strings.HasPrefix(ref, "#/") {
		parts := strings.Split(ref, "/")
		return parts[len(parts)-1]
	}

	u, err := url.Parse(ref)
	if err != nil {
		return sanitize(ref)
	}

	file := strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path))
	if u.Path == "" || u.Path == "/" {
		file = ""
	}

	if u.Host != "" {
		if file != "" {
			file = u.Host + "_" + file
		} else {
			file = u.Host
		}
	}

	var pointer string
	if u.Fragment != "" {
		segs := strings.Split(strings.Trim(u.Fragment, "/"), "/")
		last := segs[len(segs)-1]
		last = strings.ReplaceAll(last, "~1", "/")
		last = strings.ReplaceAll(last, "~0", "~")
		pointer = last
	}

	switch {
	case file != "" && pointer != "":
		return sanitize(file + "_" + pointer)
	case pointer != "":
		return sanitize(pointer)
	default:
		return sanitize(file)
	}
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	name := b.String()
	if name == "" {
		return "_"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name // avoid a leading digit
	}
	return name
}
