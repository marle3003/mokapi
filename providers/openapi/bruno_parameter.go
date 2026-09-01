package openapi

import (
	"fmt"
	"maps"
	"mokapi/export/bruno"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	"net/url"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
)

func buildParams(http *bruno.HttpDetail, params Parameters, reqPath []string) {
	newRandom := func(s *schema.Schema, additionalPath string) any {
		req := &generator.Request{
			Path:   reqPath,
			Schema: schema.ConvertToJsonSchema(s),
		}
		if additionalPath != "" {
			req.Path = append(req.Path, additionalPath)
		}

		v, err := generator.New(req)
		if err != nil {
			log.Debugf("failed to create random data for schema at %s: %v", http.Url, err)
			return ""
		}
		return v
	}

	addedQuery := false
	for _, ref := range params {
		if ref.Value == nil {
			continue
		}
		p := ref.Value
		switch p.Type {
		case ParameterHeader:
			v := newRandom(p.Schema, p.Name)
			http.Headers = append(http.Headers, bruno.HttpRequestHeader{
				Name:        p.Name,
				Value:       fmt.Sprintf("%v", v),
				Description: p.Description,
				Disabled:    !p.IsRequired(),
			})
		case ParameterPath:
			pParam := buildPathParam(p, newRandom)
			http.Params = append(http.Params, pParam)
			http.Url = strings.ReplaceAll(http.Url, fmt.Sprintf("{%s}", p.Name), fmt.Sprintf(":%s", p.Name))
		case ParameterQuery:
			qParams := buildQueryParam(p, newRandom)
			for _, q := range qParams {
				http.Params = append(http.Params, q)
				if !q.Disabled {
					if !addedQuery {
						http.Url += "?"
						addedQuery = true
					} else {
						http.Url += "&"
					}
					http.Url += fmt.Sprintf("%s=%s", q.Name, q.Value)
				}
			}
		default:
			log.Debugf("unsupported type %s for parameter %s", p.Type, p.Name)
		}
	}
}

func buildPathParam(p *Parameter, newRandom func(s *schema.Schema, name string) any) bruno.HttpRequestParam {
	param := bruno.HttpRequestParam{
		Name:        p.Name,
		Description: p.Description,
		Type:        string(ParameterPath),
	}
	r := newRandom(p.Schema, "")

	// // bruno does only encode query parameters but not path parameter
	s := serializePathParam(p.Name, p.Style, p.IsExplode(), r, p.Schema)

	param.Value = s
	return param
}

func buildQueryParam(p *Parameter, newRandom func(s *schema.Schema, name string) any) []bruno.HttpRequestParam {
	var result []bruno.HttpRequestParam

	r := newRandom(p.Schema, p.Name)
	param := bruno.HttpRequestParam{
		Name:        p.Name,
		Description: p.Description,
		Type:        string(ParameterQuery),
		Disabled:    !p.IsRequired(),
	}

	switch val := r.(type) {
	case []any:
		if p.IsExplode() {
			for _, v := range val {
				p := param
				p.Value = fmt.Sprintf("%v", v)
				result = append(result, p)
			}
		} else {
			sep := ","
			switch p.Style {
			case "spaceDelimited":
				sep = " "
			case "pipeDelimited":
				sep = "|"
			}
			for i, v := range val {
				if i > 0 {
					param.Value += sep
				}
				param.Value += fmt.Sprintf("%v", v)
			}
			result = append(result, param)
		}
	case map[string]any:
		required := map[string]bool{}
		if p.Schema != nil {
			for _, name := range p.Schema.Required {
				required[name] = true
			}
		}
		isRequired := func(name string) bool {
			_, ok := required[name]
			return ok && p.IsRequired()
		}

		var keys []string
		if p.Schema != nil && p.Schema.Properties != nil {
			keys = p.Schema.Properties.Keys()
		} else {
			keys = slices.Collect(maps.Keys(val))
		}

		if p.IsExplode() {
			for _, k := range keys {
				v, ok := val[k]
				pv := ""
				if ok {
					pv = fmt.Sprintf("%v", v)
				}
				hrp := bruno.HttpRequestParam{
					Name:     k,
					Type:     string(ParameterQuery),
					Value:    pv,
					Disabled: !isRequired(k),
				}
				if p.Schema.Properties != nil {
					prop := p.Schema.Properties.Get(k)
					if prop != nil {
						hrp.Description = prop.Description
					}
				}
				result = append(result, hrp)
			}
		} else if p.Style == "form" {
			hrp := bruno.HttpRequestParam{
				Name:        p.Name,
				Type:        string(ParameterQuery),
				Description: p.Description,
				Disabled:    !p.IsRequired(),
			}

			for i, k := range keys {
				v, ok := val[k]
				if ok {
					if i > 0 {
						hrp.Value += ","
					}
					hrp.Value += fmt.Sprintf("%v,%v", k, v)
				}
			}

			result = append(result, hrp)
		} else if p.Style == "deepObject" {
			for _, k := range keys {
				v, ok := val[k]
				pv := ""
				if ok {
					pv = fmt.Sprintf("%v", v)
				}
				hrp := bruno.HttpRequestParam{
					Name:        fmt.Sprintf("%s[%s]", p.Name, k),
					Type:        string(ParameterQuery),
					Value:       pv,
					Description: p.Description,
					Disabled:    !isRequired(k),
				}
				if p.Schema.Properties != nil {
					prop := p.Schema.Properties.Get(k)
					if prop != nil {
						hrp.Description = prop.Description
					}
				}
				result = append(result, hrp)
			}
		}
	default:
		param.Value = fmt.Sprintf("%v", val)
		result = append(result, param)
	}

	return result
}

func serializePathParam(name, style string, explode bool, value any, s *schema.Schema) string {
	switch style {
	case "matrix":
		return serializeMatrixPathParam(name, explode, value, s)
	case "label":
		return serializeLabelPathParam(explode, value, s)
	default:
		return serializeSimplePathParam(explode, value, s)
	}
}

func serializeMatrixPathParam(name string, explode bool, value any, s *schema.Schema) string {
	switch v := value.(type) {
	case []any:
		if explode {
			var sb strings.Builder
			for _, item := range v {
				sb.WriteString(fmt.Sprintf(";%s=%v",
					url.PathEscape(name), url.PathEscape(fmt.Sprint(item))),
				)
			}
			return sb.String()
		}
		return ";" + url.PathEscape(name) + "=" + joinPathArray(v, ",")
	case map[string]any:
		if explode {
			var sb strings.Builder
			for _, k := range orderedKeys(v, s) {
				val := v[k]
				sb.WriteString(fmt.Sprintf(";%s=%v", url.PathEscape(k), url.PathEscape(fmt.Sprint(val))))
			}
			return sb.String()
		}
		return ";" + name + "=" + joinPathMap(v, ",", s)
	default:
		return fmt.Sprintf(";%s=%v", url.PathEscape(name), url.PathEscape(fmt.Sprint(v)))
	}
}

func serializeLabelPathParam(explode bool, value any, s *schema.Schema) string {
	switch v := value.(type) {
	case []any:
		if explode {
			return "." + joinPathArray(v, ".")
		}
		return "." + joinPathArray(v, ",")
	case map[string]any:
		if explode {
			var sb strings.Builder
			for _, k := range orderedKeys(v, s) {
				val := v[k]
				sb.WriteString(fmt.Sprintf(".%s=%v", url.PathEscape(k), url.PathEscape(fmt.Sprint(val))))
			}
			return sb.String()
		}
		return "." + joinPathMap(v, ",", s)
	default:
		return fmt.Sprintf(".%v", url.PathEscape(fmt.Sprint(v)))
	}
}

func serializeSimplePathParam(explode bool, value any, s *schema.Schema) string {
	switch v := value.(type) {
	case []any:
		return joinPathArray(v, ",") // explode has no effect on simple arrays
	case map[string]any:
		if explode {
			var sb strings.Builder
			for _, k := range orderedKeys(v, s) {
				val := v[k]
				if sb.Len() > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(fmt.Sprintf("%s=%v", url.PathEscape(k), url.PathEscape(fmt.Sprint(val))))
			}
			return sb.String()
		}
		return joinPathMap(v, ",", s)
	default:
		return url.PathEscape(fmt.Sprint(v))
	}
}

func joinPathArray(values []any, sep string) string {
	var sb strings.Builder
	for _, value := range values {
		if sb.Len() > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(url.PathEscape(fmt.Sprint(value)))
	}
	return sb.String()
}

func joinPathMap(values map[string]any, sep string, s *schema.Schema) string {
	var sb strings.Builder
	for _, key := range orderedKeys(values, s) {
		value := values[key]
		if sb.Len() > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(url.PathEscape(key))
		sb.WriteString(sep)
		sb.WriteString(url.PathEscape(fmt.Sprint(value)))
	}
	return sb.String()
}

func orderedKeys(m map[string]any, s *schema.Schema) []string {
	var keys []string
	if s != nil && s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			k := it.Key()
			if _, ok := m[k]; ok {
				keys = append(keys, k)
			}
		}

		for k := range m {
			if _, ok := s.Properties.LinkedHashMap.Get(k); !ok {
				keys = append(keys, k)
			}
		}
	} else {
		keys = slices.Collect(maps.Keys(m))
	}

	return keys
}
