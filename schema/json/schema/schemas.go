package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/sortedmap"
)

type Schemas struct {
	sortedmap.LinkedHashMap[string, *Ref]
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
