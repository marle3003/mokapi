package schema

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/sortedmap"
	"strconv"
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

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Unmarshal(node, &r.Value)
}

func (s *SchemasRef) UnmarshalYAML(node *yaml.Node) error {
	return s.Unmarshal(node, &s.Value)
}

func (ap *AdditionalProperties) UnmarshalYAML(node *yaml.Node) error {
	ap.Allowed = true

	var err error
	if node.Kind == yaml.ScalarNode {
		ap.Allowed, err = strconv.ParseBool(node.Value)
		return err
	} else {
		return node.Decode(&ap.Ref)
	}
}
