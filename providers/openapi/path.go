package openapi

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type PathItems map[string]*PathRef

type PathRef struct {
	dynamic.Reference[*PathRef]
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

	Path   string  `yaml:"-" json:"-"`
	Status Status  `yaml:"-" json:"-"`
	Errors []Error `yaml:"-" json:"-"`
}

func (r *PathRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *PathRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	} else {
		return json.Marshal(r.Ref)
	}
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

func (p PathItems) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if p == nil {
		return nil
	}

	for name, e := range p {
		if e == nil {
			continue
		}
		if err := e.Parse(config, reader); err != nil {
			return fmt.Errorf("parse path '%v' failed: %w", name, err)
		}
		if e.Value != nil {
			e.Value.Path = name
		}
	}
	return nil
}

func (r *PathRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	return r.Value.Parse(config, reader)
}

func (p *Path) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if p == nil {
		return nil
	}

	for index, param := range p.Parameters {
		if err := param.Parse(config, reader); err != nil {
			return fmt.Errorf("parse parameter '%v' failed: %w", index, err)
		}
	}

	for method, op := range p.Operations() {
		err := op.Parse(config, reader)
		if err != nil {
			op.Status = StatusInvalid
			method = strings.ToUpper(method)
			op.Errors = append(op.Errors, Error{Message: err.Error()})
			log.
				WithField("api", getName(config)).
				WithField("method", method).
				WithField("path", p.Path).
				WithField("namespace", "http").
				Error(err)
		} else {
			op.Path = p
		}
	}

	for name, op := range p.AdditionalOperations {
		if err := op.Parse(config, reader); err != nil {
			op.Status = StatusInvalid
			name = strings.ToUpper(name)
			log.
				WithField("api", getName(config)).
				WithField("method", name).
				WithField("path", p.Path).
				WithField("namespace", "http").
				Error(err)
		} else {
			op.Path = p
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
	} else {
		r.Value.patch(patch.Value)
	}
}

func (p *Path) patch(patch *Path) {
	if p == nil || patch == nil {
		return
	}

	if len(patch.Summary) > 0 {
		p.Summary = patch.Summary
	}

	if len(patch.Description) > 0 {
		p.Description = patch.Description
	}

	if p.Get == nil {
		p.Get = patch.Get
	} else {
		p.Get.patch(patch.Get)
	}

	if p.Post == nil {
		p.Post = patch.Post
	} else {
		p.Post.patch(patch.Post)
	}

	if p.Put == nil {
		p.Put = patch.Put
	} else {
		p.Put.patch(patch.Put)
	}

	if p.Patch == nil {
		p.Patch = patch.Patch
	} else {
		p.Patch.patch(patch.Patch)
	}

	if p.Delete == nil {
		p.Delete = patch.Delete
	} else {
		p.Delete.patch(patch.Delete)
	}

	if p.Head == nil {
		p.Head = patch.Head
	} else {
		p.Head.patch(patch.Head)
	}

	if p.Options == nil {
		p.Options = patch.Options
	} else {
		p.Options.patch(patch.Options)
	}

	if p.Trace == nil {
		p.Trace = patch.Trace
	} else {
		p.Trace.patch(patch.Trace)
	}

	p.Parameters.Patch(patch.Parameters)
}
