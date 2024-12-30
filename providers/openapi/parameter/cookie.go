package parameter

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
	"strings"
)

func parseCookie(param *Parameter, r *http.Request) (*RequestParameterValue, error) {
	cookie, err := r.Cookie(param.Name)
	if err != nil || (len(cookie.Value) == 0 && param.Required) {
		if err == http.ErrNoCookie && !param.Required {
			return nil, nil
		}
		return nil, fmt.Errorf("parameter is required")
	}

	rp := &RequestParameterValue{Raw: &(cookie.Value), Value: cookie.Value}
	if param.Schema != nil {
		switch {
		case param.Schema.Value.Type.IsArray():
			rp.Value, err = parseArray(param, strings.Split(cookie.Value, ","))
		case param.Schema.Value.Type.IsObject():
			rp.Value, err = parseObject(param, cookie.Value, ",", param.IsExplode(), defaultDecode)
		default:
			rp.Value, err = p.ParseWith(cookie.Value, schema.ConvertToJsonSchema(param.Schema))
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}
