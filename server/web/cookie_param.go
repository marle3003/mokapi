package web

import (
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"net/http"
	"strings"
)

func parseCookie(p *openapi.Parameter, r *http.Request) (rp RequestParameterValue, err error) {
	switch p.Schema.Value.Type {
	case "array":
		return parseCookieArray(p, r)
	case "object":
		return parseCookieObject(p, r)
	}

	var cookie *http.Cookie
	cookie, err = r.Cookie(p.Name)
	if err != nil {
		return
	}
	rp.Raw = cookie.Value
	if len(cookie.Value) == 0 && p.Required {
		return rp, errors.Errorf("required parameter not found")
	}

	if v, err := parse(cookie.Value, p.Schema); err != nil {
		return rp, err
	} else {
		rp.Value = v
	}
	return
}

func parseCookieObject(p *openapi.Parameter, r *http.Request) (rp RequestParameterValue, err error) {
	var cookie *http.Cookie
	cookie, err = r.Cookie(p.Name)
	if err != nil {
		return
	}
	if len(cookie.Value) == 0 && p.Required {
		return rp, errors.Errorf("required parameter not found")
	}

	rp.Raw = cookie.Value
	m := make(map[string]interface{})
	rp.Value = m

	elements := strings.Split(cookie.Value, ",")
	i := 0
	for {
		if i >= len(elements) {
			break
		}
		key := elements[i]
		p, ok := p.Schema.Value.Properties.Value[key]
		if !ok {
			return rp, errors.Errorf("property '%v' not defined in schema", key)
		}
		i++
		if i >= len(elements) {
			return rp, errors.Errorf("invalid number of property pairs")
		}
		if v, err := parse(elements[i], p); err != nil {
			return rp, err
		} else {
			m[key] = v
		}
		i++
	}
	return
}

func parseCookieArray(p *openapi.Parameter, r *http.Request) (rp RequestParameterValue, err error) {
	var cookie *http.Cookie
	cookie, err = r.Cookie(p.Name)
	if err != nil {
		return
	}
	if len(cookie.Value) == 0 && p.Required {
		return rp, errors.Errorf("required parameter not found")
	}

	rp.Raw = cookie.Value
	values := make([]interface{}, 0)
	rp.Value = values

	for _, v := range strings.Split(cookie.Value, ",") {
		if i, err := parse(v, p.Schema.Value.Items); err != nil {
			return rp, err
		} else {
			values = append(values, i)
		}
	}
	return
}
