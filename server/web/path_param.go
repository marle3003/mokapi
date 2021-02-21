package web

import (
	"github.com/pkg/errors"
	"mokapi/models"
	"strings"
)

func parsePath(s string, p *models.Parameter) (interface{}, error) {
	if len(s) == 0 && p.Required {
		return nil, errors.Errorf("required parameter not found")
	}

	switch p.Style {
	case "label":
		s = s[1:]
	case "matrix":
		s = s[1:]
	}

	if p.Schema != nil {
		switch p.Schema.Type {
		case "array":
			return parsePathArray(s, p)
		case "object":
			return parsePathObject(s, p)
		}
	}

	return parse(s, p.Schema)
}

func parsePathObject(s string, p *models.Parameter) (obj map[string]interface{}, err error) {
	obj = make(map[string]interface{})
	values := strings.Split(s, ",")
	if p.Explode {
		for _, i := range values {
			kv := strings.Split(i, "=")
			if len(kv) != 2 {
				return nil, errors.Errorf("invalid format")
			}
			p, ok := p.Schema.Properties[kv[0]]
			if !ok {
				return nil, errors.Errorf("property '%v' not defined in schema", kv[0])
			}

			if v, err := parse(kv[1], p); err == nil {
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
			p, ok := p.Schema.Properties[key]
			if !ok {
				return nil, errors.Errorf("property '%v' not defined in schema", key)
			}
			i++
			if i >= len(values) {
				return nil, errors.Errorf("invalid number of property pairs")
			}
			if v, err := parse(values[i], p); err == nil {
				obj[key] = v
			} else {
				return nil, err
			}
			i++
		}
	}

	return
}

func parsePathArray(s string, p *models.Parameter) (result []interface{}, err error) {
	values := strings.Split(s, ",")

	for _, v := range values {
		if i, err := parse(v, p.Schema.Items); err != nil {
			return nil, err
		} else {
			result = append(result, i)
		}
	}
	return
}
