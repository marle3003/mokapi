package openapi

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
	"regexp"
	"strings"
)

func parsePath(param *Parameter, route string, r *http.Request) (*RequestParameterValue, error) {
	v, err := findPathValue(param, route, r)
	if err != nil {
		return nil, err
	}

	rp := &RequestParameterValue{Raw: &v, Value: v}

	switch param.Style {
	case "label":
		if v[0] != '.' {
			return nil, fmt.Errorf("expected label parameter, got at index 0: '%v'", v[0])
		}
		v = v[1:]
	case "matrix":
		if v[0] != ';' {
			return nil, fmt.Errorf("expected label parameter, got at index 0: '%v'", v[0])
		}
		v = v[1:]
	}

	if param.Schema != nil {
		switch {
		case param.Schema.Type.IsArray():
			rp.Value, err = parseArray(param, strings.Split(v, ","))
		case param.Schema.Type.IsObject():
			rp.Value, err = parseObject(param, v, ",", param.IsExplode(), defaultDecode)
		default:
			rp.Value, err = p.ParseWith(v, schema.ConvertToJsonSchema(param.Schema))
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}

func findPathValue(p *Parameter, route string, r *http.Request) (string, error) {
	requestPath := r.URL.Path
	if len(requestPath) > 1 {
		requestPath = strings.TrimRight(requestPath, "/")
	}

	servicePath, ok := r.Context().Value("servicePath").(string)
	if ok && servicePath != "/" {
		requestPath = strings.Replace(requestPath, servicePath, "", 1)
		if requestPath == "" {
			requestPath = "/"
		}
	}

	// Find all {param} names
	re := regexp.MustCompile(`\{([^}]+)\}`)
	names := re.FindAllStringSubmatch(route, -1)

	// Replace {param} with regex group
	pattern := "^" + re.ReplaceAllString(route, `([^/]+)`) + "$"
	re = regexp.MustCompile(pattern)

	// find the index for the given parameter
	index := -1
	for i, name := range names {
		if name[1] == p.Name {
			index = i
			break
		}
	}
	if index == -1 {
		return "", fmt.Errorf("path parameter %s not found in route %s", p.Name, route)
	}

	match := re.FindStringSubmatch(requestPath)
	if match == nil {
		// path parameters are always required
		return "", fmt.Errorf("url does not match route")
	}

	if len(match) < index+1 {
		return "", fmt.Errorf("parameter is required")
	}

	return match[index+1], nil
}
