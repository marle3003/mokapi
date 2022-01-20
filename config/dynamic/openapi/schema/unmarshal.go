package schema

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/sortedmap"
)

func (s *Schemas) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("not a mapping node")
	}
	s.LinkedHashMap = *sortedmap.NewLinkedHashMap()
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		val := &Ref{}
		err = value.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}

		s.Set(key, val)
	}

	return nil
}

func (s *Ref) UnmarshalYAML(node *yaml.Node) error {
	return s.Unmarshal(node, &s.Value)
}

func (s *SchemasRef) UnmarshalYAML(node *yaml.Node) error {
	return s.Unmarshal(node, &s.Value)
}
