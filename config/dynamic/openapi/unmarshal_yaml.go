package openapi

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"mokapi/media"
	"mokapi/sortedmap"
	"strconv"
)

func (r *Responses) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("not a mapping node")
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[int, *ResponseRef]{}
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		val := &ResponseRef{}
		err = value.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}

		if key == "default" {
			r.Set(0, val)
		} else {
			key, err := strconv.Atoi(key)
			if err != nil {
				log.Errorf("unable to parse http status %v", key)
				continue
			}
			r.Set(key, val)
		}
	}

	return nil
}

func (c *Content) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("not a mapping node")
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
		ct := media.ParseContentType(key)
		val.ContentType = ct
		(*c)[key] = val
	}
	return nil
}

func (r *ResponseRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *RequestBodyRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *ExampleRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *HeaderRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *NamedResponses) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *RequestBodies) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *NamedHeaders) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *Examples) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}
