package openapi

import (
	"fmt"
	"mokapi/config/dynamic"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type PathItems map[string]*PathRef

type PathRef struct {
	dynamic.Reference
	Value *Path
}

type Path struct {
	// An optional, string summary, intended to apply to all operations
	// in this path.
	Summary string

	// An optional, string description, intended to apply to all operations
	// in this path. CommonMark syntax MAY be used for rich text representation.
	Description string

	// A definition of a GET operation on this path.
	Get *Operation

	// A definition of a POST operation on this path.
	Post *Operation

	// A definition of a PUT operation on this path.
	Put *Operation

	// A definition of a PATCH operation on this path.
	Patch *Operation

	// A definition of a DELETE operation on this path.
	Delete *Operation

	// A definition of a HEAD operation on this path.
	Head *Operation

	// A definition of an OPTIONS operation on this path.
	Options *Operation

	// A definition of a TRACE operation on this path.
	Trace *Operation

	Query *Operation

	AdditionalOperations map[string]*Operation `yaml:"additionalOperations" json:"additionalOperations"`

	// A list of parameters that are applicable for all
	// the operations described under this path. These
	// parameters can be overridden at the operation level,
	// but cannot be removed there
	Parameters Parameters

	Path string `yaml:"-" json:"-"`
}

func (r *PathRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *PathRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (p *Path) Operations() map[string]*Operation {
	m := make(map[string]*Operation)
	for name, op := range p.AdditionalOperations {
		m[name] = op
	}

	if p.Get != nil {
		m[http.MethodGet] = p.Get
	}
	if p.Post != nil {
		m[http.MethodPost] = p.Post
	}
	if p.Put != nil {
		m[http.MethodPut] = p.Put
	}
	if p.Patch != nil {
		m[http.MethodPatch] = p.Patch
	}
	if p.Delete != nil {
		m[http.MethodDelete] = p.Delete
	}
	if p.Head != nil {
		m[http.MethodHead] = p.Head
	}
	if p.Options != nil {
		m[http.MethodOptions] = p.Options
	}
	if p.Trace != nil {
		m[http.MethodTrace] = p.Trace
	}
	if p.Query != nil {
		m["QUERY"] = p.Query
	}
	return m
}

func (p *Path) Operation(method string) *Operation {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodGet:
		return p.Get
	case http.MethodPost:
		return p.Post
	case http.MethodPut:
		return p.Put
	case http.MethodPatch:
		return p.Patch
	case http.MethodDelete:
		return p.Delete
	case http.MethodHead:
		return p.Head
	case http.MethodOptions:
		return p.Options
	case http.MethodTrace:
		return p.Trace
	case "QUERY":
		return p.Query
	default:
		if op, ok := p.AdditionalOperations[method]; ok {
			return op
		}
	}

	return nil
}

func (p PathItems) Resolve(token string) (interface{}, error) {
	if v, ok := p["/"+token]; ok {
		return v, nil
	}
	if v, ok := p[token]; ok {
		return v, nil
	}
	return nil, nil
}

func (p PathItems) parse(config *dynamic.Config, reader dynamic.Reader) error {
	for name, e := range p {
		if err := e.parse(name, config, reader); err != nil {
			return fmt.Errorf("parse path '%v' failed: %w", name, err)
		}
	}
	return nil
}

func (r *PathRef) parse(name string, config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	defer func() {
		if r.Value != nil {
			r.Value.Path = name
		}
	}()

	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	return r.Value.parse(config, reader)
}

func (p *Path) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if p == nil {
		return nil
	}

	for _, param := range p.Parameters {
		if err := param.Parse(config, reader); err != nil {
			return err
		}
	}

	if err := p.Get.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'GET' failed: %w", err)
	}
	if err := p.Post.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'POST' failed: %w", err)
	}
	if err := p.Put.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'PUT' failed: %w", err)
	}
	if err := p.Patch.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'PATCH' failed: %w", err)
	}
	if err := p.Delete.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'DELETE' failed: %w", err)
	}
	if err := p.Head.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'HEAD' failed: %w", err)
	}
	if err := p.Options.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'OPTIONS' failed: %w", err)
	}
	if err := p.Trace.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'TRACE' failed: %w", err)
	}
	if err := p.Query.parse(p, config, reader); err != nil {
		return fmt.Errorf("parse operation 'QUERY' failed: %w", err)
	}
	for name, op := range p.AdditionalOperations {
		if err := op.parse(p, config, reader); err != nil {
			return fmt.Errorf("parse operation '%s' failed: %w", name, err)
		}
	}

	return nil
}

func (p PathItems) patch(patch PathItems) {
	for path, v := range patch {
		if r, ok := p[path]; ok && r != nil {
			r.patch(v)
		} else {
			p[path] = v
		}
	}
}

func (r *PathRef) patch(patch *PathRef) {
	if patch == nil || patch.Value == nil {
		return
	}

	if r.Value == nil {
		r.Value = patch.Value
		return
	}

	if len(patch.Value.Summary) > 0 {
		r.Value.Summary = patch.Value.Summary
	}

	if len(patch.Value.Description) > 0 {
		r.Value.Description = patch.Value.Description
	}

	if r.Value.Get == nil {
		r.Value.Get = patch.Value.Get
	} else {
		r.Value.Get.patch(patch.Value.Get)
	}

	if r.Value.Post == nil {
		r.Value.Post = patch.Value.Post
	} else {
		r.Value.Post.patch(patch.Value.Post)
	}

	if r.Value.Put == nil {
		r.Value.Put = patch.Value.Put
	} else {
		r.Value.Put.patch(patch.Value.Put)
	}

	if r.Value.Patch == nil {
		r.Value.Patch = patch.Value.Patch
	} else {
		r.Value.Patch.patch(patch.Value.Patch)
	}

	if r.Value.Delete == nil {
		r.Value.Delete = patch.Value.Delete
	} else {
		r.Value.Delete.patch(patch.Value.Delete)
	}

	if r.Value.Head == nil {
		r.Value.Head = patch.Value.Head
	} else {
		r.Value.Head.patch(patch.Value.Head)
	}

	if r.Value.Options == nil {
		r.Value.Options = patch.Value.Options
	} else {
		r.Value.Options.patch(patch.Value.Options)
	}

	if r.Value.Trace == nil {
		r.Value.Trace = patch.Value.Trace
	} else {
		r.Value.Trace.patch(patch.Value.Trace)
	}

	r.Value.Parameters.Patch(patch.Value.Parameters)
}
