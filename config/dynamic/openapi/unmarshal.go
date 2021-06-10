package openapi

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"strconv"
)

type refProp struct {
	Ref string `yaml:"$ref" json:"$ref"`
}

func (o *Responses) UnmarshalYAML(value *yaml.Node) error {
	data := make(map[string]*ResponseRef)
	err := value.Decode(data)
	if err != nil {
		return err
	}

	*o = make(map[HttpStatus]*ResponseRef)
	for k, n := range data {
		if k == "default" {
			(*o)[Undefined] = n
		} else {
			key, err := strconv.Atoi(k)
			if err != nil {
				log.Errorf("unable to parse http status %v", k)
				continue
			}
			(*o)[HttpStatus(key)] = n
		}
	}

	return nil
}

type parameter struct {
	Name        string
	Type        ParameterLocation `yaml:"in" json:"in"`
	Schema      *SchemaRef
	Required    bool
	Description string
	Style       string
	Explode     bool
}

func (p *Parameter) UnmarshalYAML(value *yaml.Node) error {
	tmp := &parameter{Explode: true}
	err := value.Decode(tmp)
	if err != nil {
		return err
	}

	p.Name = tmp.Name
	p.Type = tmp.Type
	p.Schema = tmp.Schema
	p.Required = tmp.Required
	p.Description = tmp.Description
	p.Style = tmp.Style
	p.Explode = tmp.Explode

	return nil
}

func (s *Schemas) UnmarshalYAML(value *yaml.Node) error {
	ref := &refProp{}
	err := value.Decode(ref)
	if err != nil {
		return err
	}

	if len(ref.Ref) == 0 {
		m := make(map[string]*SchemaRef)
		err := value.Decode(m)
		if err != nil {
			return err
		}

		s.Value = m
	} else {
		s.Ref = ref.Ref
	}

	return nil
}

func (s *ResponseRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &s.Ref, &s.Value)
}

func (r *RequestBodyRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func (s *SchemaRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &s.Ref, &s.Value)
}

func (r *NamedResponses) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func (r *RequestBodies) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func (r *EndpointRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func (r *ParameterRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func unmarshalRef(node *yaml.Node, ref *string, val interface{}) error {
	r := &refProp{}
	if err := node.Decode(r); err == nil && len(r.Ref) > 0 {
		*ref = r.Ref
		return nil
	}

	return node.Decode(val)
}
