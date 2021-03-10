package web

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/openapi"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type RequestParameters map[openapi.ParameterLocation]map[string]interface{}

func newRequestParameters() RequestParameters {
	p := make(RequestParameters)
	p[openapi.PathParameter] = make(map[string]interface{})
	p[openapi.QueryParameter] = make(map[string]interface{})
	p[openapi.HeaderParameter] = make(map[string]interface{})
	p[openapi.CookieParameter] = make(map[string]interface{})
	return p
}

func parseParams(params openapi.Parameters, route string, r *http.Request) (RequestParameters, error) {
	segments := strings.Split(r.URL.Path, "/")

	path := map[string]string{}

	paramRegex := regexp.MustCompile(`\{(?P<name>.+)\}`)
	for i, segment := range strings.Split(route, "/") {
		match := paramRegex.FindStringSubmatch(segment)
		if len(match) > 1 {
			paramName := match[1]
			path[paramName] = segments[i]
		}
	}

	parameters := newRequestParameters()

	for _, ref := range params {
		if ref.Value == nil {
			continue
		}
		p := ref.Value
		var v interface{}
		var err error
		var store map[string]interface{}
		switch p.Type {
		case openapi.CookieParameter:
			v, err = parseCookie(p, r)
			store = parameters[openapi.CookieParameter]
		case openapi.PathParameter:
			if s, ok := path[p.Name]; ok {
				v, err = parsePath(s, p)
				store = parameters[openapi.PathParameter]
			} else {
				return nil, errors.Errorf("required %v paramter %q not found in request %v", p.Type, p.Name, r.URL)
			}
		case openapi.QueryParameter:
			v, err = parseQuery(p, r.URL)
			store = parameters[openapi.QueryParameter]
			//case models.HeaderParameter:
			//	value = context.Request.Header.Get(p.Name)
			//case models.PathParameter:

		}
		if err != nil && p.Required {
			return nil, errors.Wrapf(err, "parse %v parameter %q", p.Type, p.Name)
		} else if err != nil {
			log.Infof("parse %v parameter %q: %v", p.Type, p.Name, err.Error())
		}
		if store != nil {
			store[p.Name] = v
		}
	}

	return parameters, nil
}

func parse(s string, schema *openapi.SchemaRef) (interface{}, error) {
	if schema == nil {
		return s, nil
	}
	switch schema.Value.Type {
	case "string":
		return s, nil
	case "integer":
		switch schema.Value.Format {
		case "int64":
			return strconv.ParseInt(s, 10, 64)
		default:
			return strconv.Atoi(s)
		}

	case "number":
		switch schema.Value.Format {
		case "double":
			return strconv.ParseFloat(s, 64)
		default:
			v, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return nil, err
			}
			return float32(v), nil
		}
	case "boolean":
		return strconv.ParseBool(s)
	}
	return nil, errors.Errorf("unable to parse '%v'; schema type %q is not supported", s, schema.Value.Type)
}
