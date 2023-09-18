package parameter

import (
	"fmt"
	"mokapi/config/dynamic/openapi/schema"
	"net/url"
	"regexp"
	"strings"
)

func parseQuery(p *Parameter, u *url.URL) (*RequestParameterValue, error) {
	var err error

	switch {
	case p.Schema != nil && p.Schema.Value.Type == "array":
		rp := &RequestParameterValue{}
		rp.Raw, rp.Value, err = parseQueryArray(p, u)
		return rp, err
	case p.Schema != nil && p.Schema.Value.Type == "object":
		rp := &RequestParameterValue{}
		rp.Raw, rp.Value, err = parseQueryObject(p, u)
		return rp, err
	default:
		rp := &RequestParameterValue{Raw: u.Query().Get(p.Name)}
		if len(rp.Raw) == 0 {
			if p.Required {
				return nil, fmt.Errorf("parameter is required")
			} else {
				return nil, nil
			}
		}
		rp.Value, err = schema.ParseString(rp.Raw, p.Schema)
		return rp, err
	}
}

func parseQueryObject(p *Parameter, u *url.URL) (string, interface{}, error) {
	raw := ""
	if p.Style == "form" && p.IsExplode() {
		raw = u.RawQuery
		if len(raw) == 0 && p.Required {
			return "", nil, fmt.Errorf("parameter is required")
		}
		i, err := parseExplodeObject(p, raw, "&")
		return raw, i, err
	} else if p.Style == "form" {
		raw = u.Query().Get(p.Name)
		i, err := parseUnExplodeObject(p, raw, ",")
		return raw, i, err
	} else if p.Style == "deepObject" && p.IsExplode() {
		paramRegex := regexp.MustCompile(fmt.Sprintf(`%v\[(?P<name>.+)\]`, p.Name))
		obj := map[string]interface{}{}
		raw := strings.Builder{}
		for k, values := range u.Query() {
			match := paramRegex.FindStringSubmatch(k)
			name := match[1]
			prop := p.Schema.Value.Properties.Get(name)
			if prop == nil && !p.Schema.Value.IsFreeForm() && !p.Schema.Value.IsDictionary() {
				return "", nil, fmt.Errorf("property '%v' not defined in schema: %s", name, p.Schema)
			}
			s := strings.Join(values, ",")
			raw.WriteString(fmt.Sprintf("%v[%v]=%v", p.Name, name, s))
			if v, err := schema.ParseString(s, prop); err != nil {
				return "", nil, err
			} else {
				obj[name] = v
			}
		}
		return raw.String(), obj, nil

	}
	return "", nil, fmt.Errorf("unsupported style '%v', explode '%v'", p.Style, p.IsExplode())
}

func parseQueryArray(p *Parameter, u *url.URL) (string, []interface{}, error) {
	raw := ""
	sep := ","
	if p.IsExplode() {
		v := u.Query()[p.Name]
		raw = strings.Join(v, ",")
	} else {
		raw = u.Query().Get(p.Name)

		switch p.Style {
		case "spaceDelimited":
			sep = " "
		case "pipeDelimited":
			sep = "|"
		}
	}

	i, err := parseArray(p, raw, sep)
	return raw, i, err
}
