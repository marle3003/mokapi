package schema

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Schema struct {
	Type   []string  `yaml:"-" json:"-"`
	Schema []*Schema `yaml:"-" json:"-"`

	Name      string   `yaml:"name" json:"name"`
	Namespace string   `yaml:"namespace" json:"namespace"`
	Doc       string   `yaml:"doc" json:"doc"`
	Aliases   []string `yaml:"aliases" json:"aliases"`

	Fields  []Schema `yaml:"fields" json:"fields"`
	Symbols []string `yaml:"symbols" json:"symbols"`
	Items   *Schema  `yaml:"items" json:"items"`
	Values  *Schema  `yaml:"values" json:"values"`

	Order []string `yaml:"order" json:"order"`
	Size  int      `yaml:"size" json:"size"`
}

func (s *Schema) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	return nil
}

type UnmarshalError struct {
	Value interface{}
	Field string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("cannot unmarshal %v into field %v of type schema", e.Value, e.Field)
}

func (s *Schema) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		s.Type = append(s.Type, str)
		return nil
	}
	var arr []string
	err = json.Unmarshal(b, &arr)
	if err == nil {
		s.Type = append(s.Type, arr...)
		return nil
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	return s.fromMap(m)
}

func (s *Schema) UnmarshalYAML(node *yaml.Node) error {
	var str string
	err := node.Decode(&str)
	if err == nil {
		s.Type = append(s.Type, str)
		return nil
	}
	var arr []string
	err = node.Decode(&arr)
	if err == nil {
		s.Type = append(s.Type, arr...)
		return nil
	}

	m := make(map[string]interface{})
	err = node.Decode(&m)
	if err != nil {
		return err
	}

	return s.fromMap(m)
}

func (s *Schema) fromMap(m map[string]interface{}) error {
	for k, v := range m {
		switch k {
		case "type":
			switch val := v.(type) {
			case string:
				s.Type = append(s.Type, val)
			case []string:
				s.Type = append(s.Type, val...)
			case map[string]interface{}:
				nested := &Schema{}
				if err := nested.fromMap(val); err != nil {
					return err
				}
				s.Schema = append(s.Schema, nested)
			}
		case "name":
			s.Name = v.(string)
		case "namespace":
			s.Namespace = v.(string)
		case "doc":
			s.Doc = v.(string)
		case "aliases":
			s.Aliases = v.([]string)
		case "fields":
			fields := v.([]interface{})
			for _, field := range fields {
				fieldSchema := &Schema{}
				err := fieldSchema.fromMap(field.(map[string]interface{}))
				if err != nil {
					return err
				}
				s.Fields = append(s.Fields, *fieldSchema)
			}
		case "symbols":
			for _, symbol := range v.([]interface{}) {
				s.Symbols = append(s.Symbols, symbol.(string))
			}
		case "items":
			items, err := parseSchema(v)
			if err != nil {
				return err
			}
			s.Items = items
		case "values":
			values, err := parseSchema(v)
			if err != nil {
				return err
			}
			s.Values = values
		case "order":
			s.Order = v.([]string)
		case "size":
			s.Size = int(v.(float64))
		}
	}
	return nil
}

func parseSchema(v interface{}) (*Schema, error) {
	s := &Schema{}
	switch val := v.(type) {
	case string:
		s.Type = append(s.Type, val)
	case []string:
		s.Type = append(s.Type, val...)
	case map[string]interface{}:
		nested := &Schema{}
		if err := nested.fromMap(val); err != nil {
			return nil, err
		}
		s.Schema = append(s.Schema, nested)
	}
	return s, nil
}
