package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/media"
)

type Content map[string]*MediaType

//goland:noinspection GoMixedReceiverTypes
func (c *Content) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	if *c == nil {
		*c = Content{}
	}
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected openapi.Content map, got %s", token)
	}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &MediaType{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}

		ct := media.ParseContentType(key)
		val.ContentType = ct
		(*c)[key] = val
	}
}

//goland:noinspection GoMixedReceiverTypes
func (c *Content) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected openapi.Content map, got %v", value.Tag)
	}
	if *c == nil {
		*c = Content{}
	}
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		val := &MediaType{}
		err = value.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}
		if val == nil {
			return fmt.Errorf("content.%s should be object but it is nil", key)
		}
		ct := media.ParseContentType(key)
		val.ContentType = ct
		(*c)[key] = val
	}
	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (c Content) parse(config *dynamic.Config, reader dynamic.Reader) error {
	for name, mediaType := range c {
		if err := mediaType.parse(config, reader); err != nil {
			return fmt.Errorf("parse content '%v' failed: %w", name, err)
		}
	}

	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (c Content) patch(patch Content) {
	for k, v := range patch {
		if con, ok := c[k]; ok {
			con.patch(v)
		} else {
			c[k] = v
		}
	}
}
