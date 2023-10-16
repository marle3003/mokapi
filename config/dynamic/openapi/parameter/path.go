package parameter

import (
	"fmt"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
	"strings"
)

func parsePath(p *Parameter, route string, r *http.Request) (*RequestParameterValue, error) {
	path := findPathValue(p, route, r)
	if len(path) == 0 {
		// path parameters are always required
		return nil, fmt.Errorf("parameter is required")
	}

	rp := &RequestParameterValue{Raw: path, Value: path}

	switch p.Style {
	case "label":
		if path[0] != '.' {
			return nil, fmt.Errorf("expected label parameter, got at index 0: '%v'", path[0])
		}
		path = path[1:]
	case "matrix":
		if path[0] != ';' {
			return nil, fmt.Errorf("expected label parameter, got at index 0: '%v'", path[0])
		}
		path = path[1:]
	}

	var err error
	if p.Schema != nil {
		switch p.Schema.Value.Type {
		case "array":
			rp.Value, err = parseArray(p, path, ",")
		case "object":
			rp.Value, err = parseObject(p, path, ",", p.IsExplode(), defaultDecode)
		default:
			rp.Value, err = schema.ParseString(path, p.Schema)
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}

func findPathValue(p *Parameter, route string, r *http.Request) string {
	segments := strings.Split(r.URL.Path, "/")
	key := fmt.Sprintf("{%v}", p.Name)
	for i, seg := range strings.Split(route, "/") {
		if seg == key {
			return segments[i]
		}
	}
	return ""
}
