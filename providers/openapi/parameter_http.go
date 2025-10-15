package openapi

import (
	"context"
	"errors"
	"fmt"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/parser"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const requestKey = "requestParameters"

var p = parser.Parser{
	ConvertStringToNumber:  true,
	ConvertStringToBoolean: true,
}

type RequestParameters struct {
	Path        map[string]RequestParameterValue
	Query       map[string]RequestParameterValue
	Header      map[string]RequestParameterValue
	Cookie      map[string]RequestParameterValue
	QueryString *RequestParameterValue
}

type RequestParameterValue struct {
	Value any
	Raw   *string
}

type decoder func(string) (string, error)

func NewContext(ctx context.Context, rp *RequestParameters) context.Context {
	return context.WithValue(ctx, requestKey, rp)
}

func FromContext(ctx context.Context) (*RequestParameters, bool) {
	rp, ok := ctx.Value(requestKey).(*RequestParameters)
	return rp, ok
}

func FromRequest(params Parameters, route string, r *http.Request) (*RequestParameters, error) {
	parameters := &RequestParameters{
		Path:   make(map[string]RequestParameterValue),
		Query:  make(map[string]RequestParameterValue),
		Header: make(map[string]RequestParameterValue),
		Cookie: make(map[string]RequestParameterValue),
	}

	for _, ref := range params {
		if ref.Value == nil {
			continue
		}
		pv := ref.Value
		var v *RequestParameterValue
		var err error
		switch pv.Type {
		case ParameterCookie:
			v, err = parseCookie(pv, r)
			if err != nil {
				return nil, fmt.Errorf("parse cookie parameter '%v' failed: %w", pv.Name, err)
			}
			if v != nil {
				parameters.Cookie[pv.Name] = *v
			}
		case ParameterPath:
			v, err = parsePath(pv, route, r)
			if err != nil {
				if strings.Contains(route, "?") {
					return nil, fmt.Errorf("parse path parameter '%v' failed: %w. the path contains a quotation mark ('?'), which suggests query parameters are incorrectly included in the path. query parameters should be defined separately in the 'parameters' section", pv.Name, err)
				} else {
					return nil, fmt.Errorf("parse path parameter '%v' failed: %w", pv.Name, err)
				}
			}
			if v != nil {
				parameters.Path[pv.Name] = *v
			}
		case ParameterQuery:
			v, err = parseQuery(pv, r.URL)
			if err != nil {
				return nil, fmt.Errorf("parse query parameter '%v' failed: %w", pv.Name, err)
			}
			if v != nil {
				parameters.Query[pv.Name] = *v
			}
		case ParameterHeader:
			v, err = parseHeader(pv, r)
			if err != nil {
				return nil, fmt.Errorf("parse header parameter '%v' failed: %w", pv.Name, err)
			}
			if v != nil {
				parameters.Header[pv.Name] = *v
			}
		case ParameterQueryString:
			v, err = parseQueryString(pv, r)
			if err != nil {
				return nil, fmt.Errorf("parse querystring parameter '%v' failed: %w", pv.Name, err)
			}
			parameters.QueryString = v
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

func parseUnExplodeObject(param *Parameter, value, _ string) (map[string]interface{}, error) {
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

func validate(route string, params *RequestParameters) {
	re := regexp.MustCompile(`\{([^}]+)}`)
	matches := re.FindAllStringSubmatch(route, -1)

	var err error
	for _, match := range matches {
		if len(match) > 1 {
			if _, found := params.Path[match[1]]; !found {
				err = errors.Join(err, fmt.Errorf("invalid path parameter '%v'", match[1]))
			}
		}
	}

	if err != nil {
		log.Warnf("missing parameter definition for route %s: %s", route, err)
	}
}
