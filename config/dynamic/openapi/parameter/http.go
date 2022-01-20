package parameter

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
	"regexp"
	"strconv"
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
		var v RequestParameterValue
		var err error
		var store RequestParameter
		switch p.Type {
		case Cookie:
			v, err = parseCookie(p, r)
			store = parameters[Cookie]
		case Path:
			if s, ok := path[p.Name]; ok {
				v, err = parsePath(s, p)
				store = parameters[Path]
			} else {
				return nil, errors.Errorf("required path parameter %v not present", p.Name)
			}
		case Query:
			v, err = parseQuery(p, r.URL)
			store = parameters[Query]
		case Header:
			v, err = parseHeader(p, r)
			store = parameters[Header]
		}
		if err != nil && p.Required {
			return nil, errors.Wrapf(err, "%v parameter %v", p.Type, p.Name)
		} else if err != nil {
			log.Infof("%v parameter %v: %v", p.Type, p.Name, err.Error())
		}
		if store != nil {
			store[p.Name] = v
		}
	}

	return parameters, nil
}

func parse(s string, schema *schema.Ref) (interface{}, error) {
	if schema == nil {
		return s, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		for _, any := range schema.Value.AnyOf {
			if i, err := parse(s, any); err == nil {
				return i, nil
			}
		}
		return nil, errors.Errorf("unable to parse %q, not any schema matches", s)
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
