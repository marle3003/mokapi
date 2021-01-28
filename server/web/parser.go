package web

import (
	"github.com/pkg/errors"
	"mokapi/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Parameters map[string]interface{}

type RequestParameters map[models.ParameterLocation]Parameters

func newRequestParameters() RequestParameters {
	p := make(RequestParameters)
	p[models.PathParameter] = make(Parameters)
	p[models.QueryParameter] = make(Parameters)
	p[models.HeaderParameter] = make(Parameters)
	p[models.CookieParameter] = make(Parameters)
	return p
}

func parseParams(params []*models.Parameter, route string, r *http.Request) (RequestParameters, error) {
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

	for _, p := range params {
		var v interface{}
		var err error
		var store map[string]interface{}
		switch p.Location {
		case models.CookieParameter:
			v, err = parseCookie(p, r)
			store = parameters[models.CookieParameter]
		case models.PathParameter:
			if s, ok := path[p.Name]; ok {
				v, err = parsePath(s, p)
				store = parameters[models.PathParameter]
			} else {
				return nil, errors.Errorf("required %v paramter %q not found in request %v", p.Location, p.Name, r.URL)
			}
		case models.QueryParameter:
			v, err = parseQuery(p, r.URL)
			store = parameters[models.QueryParameter]
			//case models.HeaderParameter:
			//	value = context.Request.Header.Get(p.Name)
			//case models.PathParameter:

		}
		if err != nil {
			return nil, errors.Wrapf(err, "parse %v parameter %q", p.Location, p.Name)
		}
		if store != nil {
			store[p.Name] = v
		}
	}

	return parameters, nil
}

func parse(s string, schema *models.Schema) (interface{}, error) {
	switch schema.Type {
	case "string":
		return s, nil
	case "integer":
		switch schema.Format {
		case "int64":
			return strconv.ParseInt(s, 10, 64)
		default:
			return strconv.Atoi(s)
		}

	case "number":
		switch schema.Format {
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
	return nil, errors.Errorf("unable to parse '%v'", s)
}
