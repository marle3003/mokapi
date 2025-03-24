package schema

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"strconv"
)

// The gopkg.in/yaml.v3 package automatically interprets date-like strings as time.Time
// Therefore, we must use a custom parser, otherwise we will get a time.Time object that contains,
// for example, the time part in a string value, even if the value was defined as 2022-11-21

type Example struct {
	Value
}

type Value interface{}

func (e *Example) UnmarshalJSON(b []byte) error {
	return dynamic.UnmarshalJSON(b, &e.Value)
}

func (e *Example) UnmarshalYAML(node *yaml.Node) error {
	v, err := parseYamlExample(node)
	if err != nil {
		return err
	}
	e.Value = v
	return nil
}

func parseYamlExample(node *yaml.Node) (interface{}, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!int":
			return strconv.Atoi(node.Value)
		case "!!float":
			return strconv.ParseFloat(node.Value, 64)
		case "!!bool":
			return strconv.ParseBool(node.Value)
		case "!!null":
			return nil, nil
		default:
			return node.Value, nil
		}
	case yaml.MappingNode:
		m := map[string]interface{}{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			v := node.Content[i+1]
			val, err := parseYamlExample(v)
			if err != nil {
				return nil, err
			}
			m[key] = val
		}
		return m, nil
	case yaml.SequenceNode:
		var a []interface{}
		for _, v := range node.Content {
			val, err := parseYamlExample(v)
			if err != nil {
				return nil, err
			}
			a = append(a, val)
		}
		return a, nil
	default:
		return nil, fmt.Errorf("unexpected kind %v", node.Kind)
	}
}
