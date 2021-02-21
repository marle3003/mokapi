package web

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/models"
	"net/url"
	"strings"
)

func parseQuery(p *models.Parameter, u *url.URL) (interface{}, error) {
	if p.Schema != nil {
		switch p.Schema.Type {
		case "array":
			return parseQueryArray(p, u)
		case "object":
			return parseQueryObject(p, u)
		}
	}

	s := u.Query().Get(p.Name)
	if len(s) == 0 {
		if p.Required {
			return nil, errors.Errorf("required parameter not found")
		} else {
			return nil, nil
		}
	}

	return parse(s, p.Schema)
}

func parseQueryObject(p *models.Parameter, u *url.URL) (obj map[string]interface{}, err error) {
	switch s := p.Style; {
	case s == "spaceDelimited", s == "pipeDelimited":
		return nil, errors.Errorf("not supported object style '%v'", p.Style)
	case s == "deepObject" && p.Explode:
		obj = make(map[string]interface{})
		for name, prop := range p.Schema.Properties {
			s := u.Query().Get(fmt.Sprintf("%v[%v]", p.Name, name))
			if v, err := parse(s, prop); err == nil {
				obj[name] = v
			} else {
				return nil, err
			}
		}
	default:
		obj = make(map[string]interface{})
		if p.Explode {
			for name, prop := range p.Schema.Properties {
				s := u.Query().Get(name)
				if v, err := parse(s, prop); err == nil {
					obj[name] = v
				} else {
					return nil, err
				}
			}
		} else {
			s := u.Query().Get(p.Name)
			elements := strings.Split(s, ",")
			i := 0
			for {
				if i >= len(elements) {
					break
				}
				key := elements[i]
				p, ok := p.Schema.Properties[key]
				if !ok {
					return nil, errors.Errorf("property '%v' not defined in schema", key)
				}
				i++
				if i >= len(elements) {
					return nil, errors.Errorf("invalid number of property pairs")
				}
				if v, err := parse(elements[i], p); err == nil {
					obj[key] = v
				} else {
					return nil, err
				}
				i++
			}
		}
	}
	return
}

func parseQueryArray(p *models.Parameter, u *url.URL) (result []interface{}, err error) {
	var values []string
	switch s := p.Style; {
	case s == "spaceDelimited" && !p.Explode:
		v, ok := u.Query()[p.Name]
		if !ok && p.Required {
			return nil, errors.Errorf("required parameter not found")
		}
		values = strings.Split(v[0], " ")
	case s == "pipeDelimited" && !p.Explode:
		v, ok := u.Query()[p.Name]
		if !ok && p.Required {
			return nil, errors.Errorf("required parameter not found")
		}
		values = strings.Split(v[0], "|")
	case s == "deepObject":
	default:
		if p.Explode {
			var ok bool
			values, ok = u.Query()[p.Name]
			if !ok && p.Required {
				return nil, errors.Errorf("required parameter not found")
			}
		} else {
			s := u.Query().Get(p.Name)
			if len(s) == 0 && p.Required {
				return nil, errors.Errorf("required parameter not found")
			}
			values = strings.Split(s, ",")
		}

	}

	for _, v := range values {
		if i, err := parse(v, p.Schema.Items); err != nil {
			return nil, err
		} else {
			result = append(result, i)
		}
	}
	return
}
