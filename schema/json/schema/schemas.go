package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/sortedmap"
)

type Schemas struct {
	sortedmap.LinkedHashMap[string, *Schema]
}

func (s *Schemas) Get(name string) *Schema {
	if s == nil {
		return nil
	}
	r, _ := s.LinkedHashMap.Get(name)
	return r
}

func (s *Schemas) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("not a mapping node")
	}
	s.LinkedHashMap = sortedmap.LinkedHashMap[string, *Schema]{}
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		val := &Schema{}
		err = value.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}

		s.Set(key, val)
	}

	return nil
}

func (s *Schemas) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected openapi.Responses map, got %s", token)
	}
	s.LinkedHashMap = sortedmap.LinkedHashMap[string, *Schema]{}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &Schema{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}
		s.Set(key, val)
	}
}

func (s *Schemas) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if s == nil {
		return nil
	}

	for it := s.Iter(); it.Next(); {
		if err := it.Value().Parse(config, reader); err != nil {
			inner := errors.Unwrap(err)
			return fmt.Errorf("parse schema '%v' failed: %w", it.Key(), inner)
		}
	}

	return nil
}

func (s *Schemas) Resolve(token string) (interface{}, error) {
	i := s.Get(token)
	if i == nil {
		return nil, fmt.Errorf("unable to resolve %v", token)
	}
	return i, nil
}
