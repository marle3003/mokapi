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
	data := make(map[string]*Response)
	value.Decode(data)

	*o = make(map[HttpStatus]*Response)
	for k, n := range data {
		key, err := strconv.Atoi(k)
		if err != nil {
			log.Errorf("unable to parse http status %v", k)
			continue
		}
		(*o)[HttpStatus(key)] = n
	}

	return nil
}

type parameter struct {
	Name        string
	Type        ParameterLocation `yaml:"in" json:"in"`
	Schema      *Schema
	Required    bool
	Description string
	Style       string
	Explode     bool
}

func (p *Parameter) UnmarshalYAML(value *yaml.Node) error {
	tmp := &parameter{Explode: true}
	value.Decode(tmp)

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
	value.Decode(ref)
	if len(ref.Ref) == 0 {
		m := make(map[string]*SchemaRef)
		value.Decode(m)
		s.Value = m
	} else {
		s.Ref = ref.Ref
	}

	return nil
}
