package parameter

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
	"strings"
)

const requestKey = "requestParameters"

type RequestParameters map[Location]RequestParameter

func newRequestParameters() RequestParameters {
	p := make(RequestParameters)
	p[Path] = make(RequestParameter)
	p[Query] = make(RequestParameter)
	p[Header] = make(RequestParameter)
	p[Cookie] = make(RequestParameter)
	return p
}

type RequestParameter map[string]RequestParameterValue

type RequestParameterValue struct {
	Value interface{}
	Raw   string
}

func NewContext(ctx context.Context, rp RequestParameters) context.Context {
	return context.WithValue(ctx, requestKey, rp)
}

func FromContext(ctx context.Context) (RequestParameters, bool) {
	rp, ok := ctx.Value(requestKey).(RequestParameters)
	return rp, ok
}

func FromRequest(params Parameters, route string, r *http.Request) (RequestParameters, error) {
	parameters := newRequestParameters()

	for _, ref := range params {
		if ref.Value == nil {
			continue
		}
		p := ref.Value
		var v *RequestParameterValue
		var err error
		var store RequestParameter
		switch p.Type {
		case Cookie:
			v, err = parseCookie(p, r)
			store = parameters[Cookie]
			if err != nil {
				err = fmt.Errorf("parse cookie parameter '%v' failed: %w", p.Name, err)
			}
		case Path:
			v, err = parsePath(p, route, r)
			store = parameters[Path]
			if err != nil {
				err = fmt.Errorf("parse path parameter '%v' failed: %w", p.Name, err)
			}
		case Query:
			v, err = parseQuery(p, r.URL)
			store = parameters[Query]
			if err != nil {
				err = fmt.Errorf("parse query parameter '%v' failed: %w", p.Name, err)
			}
		case Header:
			v, err = parseHeader(p, r)
			store = parameters[Header]
			if err != nil {
				err = fmt.Errorf("parse header parameter '%v' failed: %w", p.Name, err)
			}
		}
		if err != nil {
			return nil, err
		}
		if store != nil && v != nil {
			store[p.Name] = *v
		}
	}

	return parameters, nil
}

func parseObject(p *Parameter, value string, separator string, explode bool) (map[string]interface{}, error) {
	if explode {
		return parseExplodeObject(p, value, separator)
	} else {
		return parseUnExplodeObject(p, value, separator)
	}
}

func parseExplodeObject(p *Parameter, value, separator string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	values := strings.Split(value, separator)
	for _, i := range values {
		kv := strings.Split(i, "=")
		if len(kv) != 2 {
			return nil, errors.Errorf("invalid format")
		}
		prop := p.Schema.Value.Properties.Get(kv[0])
		if prop == nil && !p.Schema.Value.IsFreeForm() && !p.Schema.Value.IsDictionary() {
			return nil, fmt.Errorf("property '%v' not defined in schema: %s", kv[0], p.Schema)
		}

		if v, err := schema.ParseString(kv[1], prop); err == nil {
			m[kv[0]] = v
		} else {
			return nil, err
		}
	}
	return m, nil
}

func parseUnExplodeObject(p *Parameter, value, separator string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	elements := strings.Split(value, ",")
	i := 0
	for {
		if i >= len(elements) {
			break
		}
		key := elements[i]
		i++

		p := p.Schema.Value.Properties.Get(key)
		if p == nil {
			continue
		}
		if i >= len(elements) {
			return nil, fmt.Errorf("invalid number of property pairs")
		}
		if v, err := schema.ParseString(elements[i], p); err != nil {
			return nil, fmt.Errorf("parse property '%v' failed: %w", key, err)
		} else {
			m[key] = v
		}
		i++
	}
	return m, nil
}

func parseArray(p *Parameter, value string, separator string) ([]interface{}, error) {
	values := make([]interface{}, 0)

	for _, v := range strings.Split(value, separator) {
		if i, err := schema.ParseString(v, p.Schema.Value.Items); err != nil {
			return nil, err
		} else {
			values = append(values, i)
		}
	}

	return values, nil
}
