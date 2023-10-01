package parameter

import (
	"fmt"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
)

func parseCookie(p *Parameter, r *http.Request) (*RequestParameterValue, error) {
	cookie, err := r.Cookie(p.Name)
	if err != nil || (len(cookie.Value) == 0 && p.Required) {
		if err == http.ErrNoCookie && !p.Required {
			return nil, nil
		}
		return nil, fmt.Errorf("parameter is required")
	}

	rp := &RequestParameterValue{Raw: cookie.Value, Value: cookie.Value}
	if p.Schema != nil {
		switch p.Schema.Value.Type {
		case "array":
			rp.Value, err = parseArray(p, cookie.Value, ",")
		case "object":
			rp.Value, err = parseObject(p, cookie.Value, ",", p.IsExplode(), defaultDecode)
		default:
			rp.Value, err = schema.ParseString(cookie.Value, p.Schema)
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}
