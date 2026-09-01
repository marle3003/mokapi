package openapi

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var paramTokenRegex = regexp.MustCompile(`{([^}]+)}`)

func findPathValues(params []*Parameter, route string, r *http.Request) (map[string]RequestParameterValue, error) {
	requestPath := r.URL.RawPath
	if requestPath == "" {
		requestPath = r.URL.Path
	}
	if len(requestPath) > 1 {
		requestPath = strings.TrimRight(requestPath, "/")
	}

	servicePath, ok := r.Context().Value("servicePath").(string)
	if ok && servicePath != "/" {
		requestPath = strings.TrimPrefix(requestPath, servicePath)
		if requestPath == "" {
			requestPath = "/"
		}
	}

	routeSegs, reqSegs, err := pathSegments(route, requestPath)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]*Parameter, len(params))
	for _, p := range params {
		byName[p.Name] = p
	}

	values := map[string]RequestParameterValue{}

	for i, routeSeg := range routeSegs {
		names := tokensInSegment(routeSeg)
		if len(names) == 0 {
			if routeSeg != reqSegs[i] {
				return nil, fmt.Errorf("literal segment mismatch: expected %q, got %q", routeSeg, reqSegs[i])
			}
			continue
		}

		rawSeg := reqSegs[i]

		if len(names) == 1 {
			param := byName[names[0]]
			if param == nil {
				continue
			}

			err = parseSimplePathParam([]*Parameter{param}, routeSeg, rawSeg, values)
			if err != nil {
				return nil, err
			}
			continue
		}

		style := "simple"
		for j, name := range names {
			param := byName[name]
			if param == nil {
				return nil, fmt.Errorf("specification for path parameter %q not found", name)
			}
			if j == 0 {
				style = param.Style
			} else if style != param.Style {
				return nil, fmt.Errorf("parameter style mismatch for %s: expected %q, got %q", strings.Join(names, ","), style, param.Style)
			}
			if param.IsExplode() {
				return nil, fmt.Errorf("exploded parameter not supported for multiple parameters in same segment: %s", strings.Join(names, ","))
			}
		}

		var sep string
		switch style {
		case "matrix":
			sep = ";"
		case "label":
			sep = "."
		default:
			paramSegments := make([]*Parameter, 0, len(names))
			for _, name := range names {
				paramSegments = append(paramSegments, byName[name])
			}
			err = parseSimplePathParam(paramSegments, routeSeg, rawSeg, values)
			if err != nil {
				return nil, err
			}
			continue
		}
		items := strings.Split(rawSeg, sep)
		if len(items) != len(names)+1 {
			return nil, fmt.Errorf("segment %q: expected %d parameters, got %d chunks", rawSeg, len(names), len(items)-1)
		}
		if items[0] != "" {
			return nil, fmt.Errorf("segment %q does not start with expected separator %q", rawSeg, sep)
		}

		for j := 1; j < len(items); j++ {
			param := byName[names[j-1]]
			raw := items[j]
			v, err := decodeStyledValue(param, raw)
			if err != nil {
				return nil, fmt.Errorf("param %q: %w", param.Name, err)
			}
			values[param.Name] = RequestParameterValue{
				Value: v,
				Raw:   &raw,
			}
		}
	}

	return values, nil
}

func parseSimplePathParam(params []*Parameter, routeSeg string, rawSeg string, values map[string]RequestParameterValue) error {
	pattern := buildSegmentPattern(routeSeg)
	match := regexp.MustCompile(pattern).FindStringSubmatch(rawSeg)
	if match == nil {
		return fmt.Errorf("segment %q does not match route %q", rawSeg, routeSeg)
	}

	for i, param := range params {
		val, err := decodeStyledValue(param, match[i+1])
		if err != nil {
			return fmt.Errorf("param %q: %w", param.Name, err)
		}
		values[param.Name] = RequestParameterValue{
			Value: val,
			Raw:   &rawSeg,
		}
	}

	return nil
}

func pathSegments(route, path string) ([]string, []string, error) {
	routeSegs := strings.Split(strings.Trim(route, "/"), "/")
	reqSegs := strings.Split(strings.Trim(path, "/"), "/")
	if len(routeSegs) != len(reqSegs) {
		return nil, nil, fmt.Errorf("route and path must have the same segment length")
	}
	return routeSegs, reqSegs, nil
}

func tokensInSegment(tmplSeg string) []string {
	matches := paramTokenRegex.FindAllStringSubmatch(tmplSeg, -1)
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = m[1]
	}
	return names
}

func decodeStyledValue(param *Parameter, raw string) (any, error) {
	value, err := url.PathUnescape(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape path parameter %q: %w", param.Name, err)
	}

	sep := ","
	switch param.Style {
	case "matrix":
		value = strings.TrimPrefix(value, ";")
		if !(param.IsExplode() && param.Schema.Type.IsObject()) {
			value = parsePathMatrix(param, value)
		} else {
			sep = ";"
		}
	case "label":
		value = parsePathLabel(value)
	}

	if param.Schema != nil {
		switch {
		case param.Schema.Type.IsArray():
			return parseArray(param, strings.Split(value, sep))
		case param.Schema.Type.IsObject():
			return parseObject(param, value, sep, param.IsExplode(), defaultDecode)
		}
	}
	return p.ParseWith(value, schema.ConvertToJsonSchema(param.Schema))
}

func parsePathMatrix(param *Parameter, s string) string {
	var rex = regexp.MustCompile(fmt.Sprintf("(%s)=([^;]*)", regexp.QuoteMeta(param.Name)))
	data := rex.FindAllStringSubmatch(s, -1)

	result := ""
	for i, d := range data {
		if i > 0 {
			result += ","
		}
		result += d[2]
	}

	return result
}

func parsePathLabel(s string) string {
	s = strings.TrimPrefix(s, ".")
	split := strings.Split(s, ".")

	result := ""
	for i, v := range split {
		if i > 0 {
			result += ","
		}
		result += v
	}

	return result
}

func buildSegmentPattern(routeSeg string) string {
	var sb strings.Builder
	sb.WriteString("^")

	lastEnd := 0
	matches := paramTokenRegex.FindAllStringIndex(routeSeg, -1)

	for _, m := range matches {
		start, end := m[0], m[1]
		literal := routeSeg[lastEnd:start]
		sb.WriteString(regexp.QuoteMeta(literal))
		sb.WriteString(`(.*)`)
		lastEnd = end
	}
	sb.WriteString(regexp.QuoteMeta(routeSeg[lastEnd:]))
	sb.WriteString("$")

	return sb.String()
}
