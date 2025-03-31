package parameter

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/parser"
	"net/http"
	"regexp"
	"strings"
)

const requestKey = "requestParameters"

var p = parser.Parser{
	ConvertStringToNumber:  true,
	ConvertStringToBoolean: true,
}

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
	Raw   *string
}

type decoder func(string) (string, error)

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
				if strings.Contains(route, "?") {
					err = fmt.Errorf("parse path parameter '%v' failed: %w. the path contains a quotation mark ('?'), which suggests query parameters are incorrectly included in the path. query parameters should be defined separately in the 'parameters' section", p.Name, err)
				} else {
					err = fmt.Errorf("parse path parameter '%v' failed: %w", p.Name, err)
				}
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

	validate(route, parameters)

	return parameters, nil
}

func parseObject(p *Parameter, value string, separator string, explode bool, decode decoder) (map[string]interface{}, error) {
	if explode {
		return parseExplodeObject(p, value, separator, decode)
	} else {
		return parseUnExplodeObject(p, value, separator)
	}
}

func parseExplodeObject(param *Parameter, value, separator string, decode decoder) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	values := strings.Split(value, separator)
	for _, i := range values {
		kv := strings.Split(i, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid format")
		}
		key, err := decode(kv[0])
		if err != nil {
			return nil, err
		}
		val, err := decode(kv[1])
		if err != nil {
			return nil, err
		}
		prop := param.Schema.Properties.Get(key)
		if prop == nil && !param.Schema.IsFreeForm() && !param.Schema.IsDictionary() {
			return nil, fmt.Errorf("property '%v' not defined in schema: %s", kv[0], param.Schema)
		}

		if v, err := p.ParseWith(val, schema.ConvertToJsonSchema(prop)); err == nil {
			m[key] = v
		} else {
			return nil, err
		}
	}
	return m, nil
}

func parseUnExplodeObject(param *Parameter, value, separator string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	elements := strings.Split(value, ",")
	i := 0
	for {
		if i >= len(elements) {
			break
		}
		key := elements[i]
		i++

		prop := param.Schema.Properties.Get(key)
		if prop == nil {
			continue
		}
		if i >= len(elements) {
			return nil, fmt.Errorf("invalid number of property pairs")
		}
		if v, err := p.ParseWith(elements[i], schema.ConvertToJsonSchema(prop)); err != nil {
			return nil, fmt.Errorf("parse property '%v' failed: %w", key, err)
		} else {
			m[key] = v
		}
		i++
	}
	return m, nil
}

func parseArray(param *Parameter, value []string) (interface{}, error) {
	return p.ParseWith(value, schema.ConvertToJsonSchema(param.Schema))
}

func defaultDecode(s string) (string, error) {
	return s, nil
}

func validate(route string, params RequestParameters) {
	re := regexp.MustCompile(`\{([^}]+)}`)
	matches := re.FindAllStringSubmatch(route, -1)

	var err error
	for _, match := range matches {
		if len(match) > 1 {
			if _, found := params[Path][match[1]]; !found {
				err = errors.Join(err, fmt.Errorf("invalid path parameter '%v'", match[1]))
			}
		}
	}

	if err != nil {
		log.Warnf("missing parameter definition for route %s: %s", route, err)
	}
}
