package parameter

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi/schema"
	"strings"
)

func parsePath(s string, p *Parameter) (rp RequestParameterValue, err error) {
	if len(s) == 0 && p.Required {
		return rp, fmt.Errorf("path parameter '%v' is required", p.Name)
	}

	rp.Raw = s

	switch p.Style {
	case "label":
		s = s[1:]
	case "matrix":
		s = s[1:]
	}

	var v interface{}
	if p.Schema != nil {
		switch p.Schema.Value.Type {
		case "array":
			v, err = parsePathArray(s, p)
		case "object":
			v, err = parsePathObject(s, p)
		default:
			v, err = schema.ParseString(s, p.Schema)
		}
	}

	if err != nil {
		return rp, err
	} else {
		rp.Value = v
	}

	return
}

func parsePathObject(s string, p *Parameter) (obj map[string]interface{}, err error) {
	obj = make(map[string]interface{})
	values := strings.Split(s, ",")
	if p.Explode {
		for _, i := range values {
			kv := strings.Split(i, "=")
			if len(kv) != 2 {
				return nil, errors.Errorf("invalid format")
			}
			p := p.Schema.Value.Properties.Get(kv[0])
			if p == nil {
				return nil, errors.Errorf("property '%v' not defined in schema", kv[0])
			}

			if v, err := schema.ParseString(kv[1], p); err == nil {
				obj[kv[0]] = v
			} else {
				return nil, err
			}
		}
	} else {
		i := 0
		for {
			if i >= len(values) {
				break
			}
			key := values[i]
			p := p.Schema.Value.Properties.Get(key)
			if p == nil {
				return nil, errors.Errorf("property '%v' not defined in schema", key)
			}
			i++
			if i >= len(values) {
				return nil, errors.Errorf("invalid number of property pairs")
			}
			if v, err := schema.ParseString(values[i], p); err == nil {
				obj[key] = v
			} else {
				return nil, err
			}
			i++
		}
	}

	return
}

func parsePathArray(s string, p *Parameter) (result []interface{}, err error) {
	values := strings.Split(s, ",")

	for _, v := range values {
		if i, err := schema.ParseString(v, p.Schema.Value.Items); err != nil {
			return nil, err
		} else {
			result = append(result, i)
		}
	}
	return
}
