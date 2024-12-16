package parameter

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
	"strings"
)

func parsePath(param *Parameter, route string, r *http.Request) (*RequestParameterValue, error) {
	path := findPathValue(param, route, r)
	if len(path) == 0 {
		// path parameters are always required
		return nil, fmt.Errorf("parameter is required")
	}

	rp := &RequestParameterValue{Raw: &path, Value: path}

	switch param.Style {
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
	if param.Schema != nil {
		switch {
		case param.Schema.Value.Type.IsArray():
			rp.Value, err = parseArray(param, strings.Split(path, ","))
		case param.Schema.Value.Type.IsObject():
			rp.Value, err = parseObject(param, path, ",", param.IsExplode(), defaultDecode)
		default:
			rp.Value, err = p.Parse(path, schema.ConvertToJsonSchema(param.Schema))
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
