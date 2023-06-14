package parameter

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
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
			return nil, fmt.Errorf("%v: %v parameter %v", err, p.Type, p.Name)
		} else if err != nil {
			log.Infof("%v parameter %v: %v", p.Type, p.Name, err.Error())
		}
		if store != nil {
			store[p.Name] = v
		}
	}

	return parameters, nil
}
