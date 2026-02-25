package openapi

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
	"strings"
)

func parseCookie(param *Parameter, r *http.Request) (*RequestParameterValue, error) {
	cookie, err := r.Cookie(param.Name)
	if err != nil || len(cookie.Value) == 0 {
		if param.Required {
			return nil, fmt.Errorf("parameter is required")
		}
		if param.Schema != nil && param.Schema.Default != nil {
			return &RequestParameterValue{Value: param.Schema.Default}, nil
		}
		return nil, nil
	}

	rp := &RequestParameterValue{Raw: &(cookie.Value), Value: cookie.Value}
	if param.Schema != nil {
		switch {
		case param.Schema.Type.IsArray():
			rp.Value, err = parseArray(param, strings.Split(cookie.Value, ","))
		case param.Schema.Type.IsObject():
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
