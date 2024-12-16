package parameter

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/url"
	"regexp"
	"strings"
)

func parseQuery(param *Parameter, u *url.URL) (*RequestParameterValue, error) {
	var err error

	switch {
	case param.Schema != nil && param.Schema.Value.Type.IsArray():
		rp := &RequestParameterValue{}
		rp.Raw, rp.Value, err = parseQueryArray(param, u)
		return rp, err
	case param.Schema != nil && param.Schema.Value.Type.IsObject():
		rp := &RequestParameterValue{}
		var raw string
		raw, rp.Value, err = parseQueryObject(param, u)
		rp.Raw = &raw
		return rp, err
	default:
		if !u.Query().Has(param.Name) {
			if param.Required {
				return nil, fmt.Errorf("parameter is required")
			}
			return &RequestParameterValue{}, err
		}
		raw := u.Query().Get(param.Name)
		rp := &RequestParameterValue{Raw: &raw}
		if len(*rp.Raw) == 0 {
			if param.Required {
				return nil, fmt.Errorf("parameter is required")
			} else {
				return nil, nil
			}
		}
		rp.Value, err = p.Parse(*rp.Raw, schema.ConvertToJsonSchema(param.Schema))
		return rp, err
	}
}

func parseQueryObject(param *Parameter, u *url.URL) (string, interface{}, error) {
	if param.Style == "form" && param.IsExplode() {
		raw := u.RawQuery
		if len(raw) == 0 && param.Required {
			return "", nil, fmt.Errorf("parameter is required")
		}
		i, err := parseExplodeObject(param, raw, "&", url.QueryUnescape)
		return raw, i, err
	} else if param.Style == "form" {
		raw := u.Query().Get(param.Name)
		if len(raw) == 0 && param.Required {
			return "", nil, fmt.Errorf("parameter is required")
		}
		i, err := parseUnExplodeObject(param, raw, ",")
		return raw, i, err
	} else if param.Style == "deepObject" && param.IsExplode() {
		paramRegex := regexp.MustCompile(fmt.Sprintf(`%v\[(?P<name>.+)\]`, param.Name))
		obj := map[string]interface{}{}
		raw := strings.Builder{}

		for k, values := range u.Query() {
			match := paramRegex.FindStringSubmatch(k)
			name := match[1]
			prop := param.Schema.Value.Properties.Get(name)
			if prop == nil && !param.Schema.Value.IsFreeForm() && !param.Schema.Value.IsDictionary() {
				return "", nil, fmt.Errorf("property '%v' not defined in schema: %s", name, param.Schema)
			}
			s := strings.Join(values, ",")
			raw.WriteString(fmt.Sprintf("%v[%v]=%v", param.Name, name, s))
			if v, err := p.Parse(s, schema.ConvertToJsonSchema(prop)); err != nil {
				return "", nil, err
			} else {
				obj[name] = v
			}
		}
		if len(raw.String()) == 0 && param.Required {
			return "", nil, fmt.Errorf("parameter is required")
		}
		return raw.String(), obj, nil

	}
	return "", nil, fmt.Errorf("unsupported style '%v', explode '%v'", param.Style, param.IsExplode())
}

func parseQueryArray(p *Parameter, u *url.URL) (*string, interface{}, error) {
	var raw string
	var values []string
	if u.Query().Has(p.Name) {
		if p.IsExplode() {
			values = u.Query()[p.Name]
			raw = strings.Join(values, ",")
		} else {
			raw = u.Query().Get(p.Name)

			switch p.Style {
			case "spaceDelimited":
				values = strings.Split(raw, " ")
			case "pipeDelimited":
				values = strings.Split(raw, "|")
			default:
				values = strings.Split(raw, ",")
			}
		}

	}
	i, err := parseArray(p, values)
	return &raw, i, err
}
