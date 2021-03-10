package web

import (
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"net/http"
	"strings"
)

func parseCookie(p *openapi.Parameter, r *http.Request) (interface{}, error) {
	switch p.Schema.Value.Type {
	case "array":
		return parseCookieArray(p, r)
	case "object":
		return parseCookieObject(p, r)
	}

	c, err := r.Cookie(p.Name)
	if err != nil {
		return nil, err
	}
	if len(c.Value) == 0 && p.Required {
		return nil, errors.Errorf("required parameter not found")
	}

	return parse(c.Value, p.Schema)
}

func parseCookieObject(p *openapi.Parameter, r *http.Request) (obj map[string]interface{}, err error) {
	c, err := r.Cookie(p.Name)
	if err != nil {
		return nil, err
	}
	if len(c.Value) == 0 && p.Required {
		return nil, errors.Errorf("required parameter not found")
	}

	elements := strings.Split(c.Value, ",")
	i := 0
	for {
		if i >= len(elements) {
			break
		}
		key := elements[i]
		p, ok := p.Schema.Value.Properties.Value[key]
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
	return
}

func parseCookieArray(p *openapi.Parameter, r *http.Request) (result []interface{}, err error) {
	c, err := r.Cookie(p.Name)
	if err != nil {
		return nil, err
	}
	if len(c.Value) == 0 && p.Required {
		return nil, errors.Errorf("required parameter not found")
	}

	values := strings.Split(c.Value, ",")

	for _, v := range values {
		if i, err := parse(v, p.Schema.Value.Items); err != nil {
			return nil, err
		} else {
			result = append(result, i)
		}
	}
	return
}
