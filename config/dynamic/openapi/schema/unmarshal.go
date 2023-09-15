package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"mokapi/sortedmap"
)

func (s *Schemas) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected openapi.Responses map, got %s", token)
	}
	s.LinkedHashMap = sortedmap.LinkedHashMap[string, *Ref]{}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &Ref{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}
		s.Set(key, val)
	}
}

func (s *Schemas) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("not a mapping node")
	}
	s.LinkedHashMap = sortedmap.LinkedHashMap[string, *Ref]{}
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

func (r *Ref) UnmarshalJSON(b []byte) error {
	return r.UnmarshalJson(b, &r.Value)
}

func (ap *AdditionalProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		var b bool
		err := node.Decode(&b)
		if err != nil {
			return err
		}
		ap.Forbidden = !b
		return err
	} else {
		return node.Decode(&ap.Ref)
	}
}

func (ap *AdditionalProperties) UnmarshalJSON(b []byte) error {
	var allowed bool
	err := json.Unmarshal(b, &allowed)
	if err == nil {
		ap.Forbidden = !allowed
		return nil
	}
	return json.Unmarshal(b, &ap.Ref)
}
