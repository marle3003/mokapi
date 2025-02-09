package schema

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"strings"
)

var table = map[string]*Schema{}

type Schema struct {
	Type       []interface{}      `yaml:"-" json:"-"`
	NamedTypes map[string]*Schema `yaml:"-" json:"-"`

	Name      string   `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace string   `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Doc       string   `yaml:"doc,omitempty" json:"doc,omitempty"`
	Aliases   []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`

	Fields  []*Schema `yaml:"fields,omitempty" json:"fields,omitempty"`
	Symbols []string  `yaml:"symbols,omitempty" json:"symbols,omitempty"`
	Items   *Schema   `yaml:"items,omitempty" json:"items,omitempty"`
	Values  *Schema   `yaml:"values,omitempty" json:"values,omitempty"`

	Order []string `yaml:"order,omitempty" json:"order,omitempty"`
	Size  int      `yaml:"size,omitempty" json:"size,omitempty"`

	fullname  string
	namespace string
}

func (s *Schema) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	ns := s.Namespace
	name := s.Name
	if strings.Contains(name, ".") {
		i := strings.LastIndex(name, ".")
		ns = name[:i]
		name = name[i+1:]
	}
	if ns == "" {
		ns = config.Scope.Name()
	} else {
		config.OpenScope(ns)
		defer config.CloseScope()
	}
	s.namespace = ns

	if s.Name != "" {
		if ns != "" {
			s.fullname = fmt.Sprintf("%s.%s", ns, name)
		} else {
			s.fullname = name
		}
		table[s.fullname] = s
	}

	for i, t := range s.Type {
		switch v := t.(type) {
		case string:
			nt, err := config.Scope.GetDynamic(v)
			if err == nil {
				s.Type[i] = nt
			}
		case *Schema:
			err := v.Parse(config, reader)
			if err != nil {
				return err
			}
		}
	}

	for _, f := range s.Fields {
		err := f.Parse(config, reader)
		if err != nil {
			return err
		}
	}

	if s.Items != nil {
		return s.Items.Parse(config, reader)
	}

	if s.Values != nil {
		return s.Values.Parse(config, reader)
	}

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
		for _, t := range arr {
			s.Type = append(s.Type, t)
		}
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
		for _, t := range arr {
			s.Type = append(s.Type, t)
		}
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
				for _, t := range val {
					s.Type = append(s.Type, t)
				}
			case []interface{}:
				for _, t := range val {
					switch u := t.(type) {
					case string:
						s.Type = append(s.Type, u)
					case map[string]interface{}:
						nested := &Schema{}
						if err := nested.fromMap(u); err != nil {
							return err
						}
						s.Type = append(s.Type, nested)
					}
				}
			case map[string]interface{}:
				nested := &Schema{}
				if err := nested.fromMap(val); err != nil {
					return err
				}
				s.Type = append(s.Type, nested)
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
				s.Fields = append(s.Fields, fieldSchema)
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
		for _, t := range val {
			s.Type = append(s.Type, t)
		}
	case map[string]interface{}:
		nested := &Schema{}
		if err := nested.fromMap(val); err != nil {
			return nil, err
		}
		s.Type = append(s.Type, nested)
	}
	return s, nil
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	type alias Schema
	a := alias(*s)
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	var typeValue strings.Builder
	if len(s.Type) == 1 {
		single, err := json.Marshal(s.Type[0])
		if err != nil {
			return nil, err
		}
		typeValue.WriteString(fmt.Sprintf(`"type":%s`, single))
	} else if len(s.Type) > 1 {
		list, err := json.Marshal(s.Type)
		if err != nil {
			return nil, err
		}
		typeValue.WriteString(fmt.Sprintf(`"type":%s`, list))
	}

	content := b[1 : len(b)-1]
	if len(content) > 0 {
		return []byte(fmt.Sprintf(`{%s,%s}`, typeValue.String(), content)), nil
	}
	return []byte(fmt.Sprintf(`{%s}`, typeValue.String())), nil
}

func getFullname(s *Schema, typeName string) string {
	if !strings.Contains(typeName, ".") && s.namespace != "" {
		typeName = fmt.Sprintf("%s.%s", s.namespace, typeName)
	}
	return typeName
}

func (s *Schema) String() string {
	var sb strings.Builder
	if len(s.fullname) > 0 {
		sb.WriteString(fmt.Sprintf("%s: ", s.fullname))
	}

	sb.WriteString(fmt.Sprintf("types: %v", s.Type))

	return sb.String()
}
