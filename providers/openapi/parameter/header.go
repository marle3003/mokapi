package parameter

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
	"strings"
)

func parseHeader(param *Parameter, r *http.Request) (*RequestParameterValue, error) {
	header := r.Header.Get(param.Name)

	if len(header) == 0 {
		if param.Required {
			return nil, fmt.Errorf("parameter is required")
		}
		return nil, nil
	}

	rp := &RequestParameterValue{Raw: &header, Value: header}
	var err error
	if param.Schema != nil {
		switch {
		case param.Schema.Value.Type.IsArray():
			rp.Value, err = parseArray(param, strings.Split(header, ","))
		case param.Schema.Value.Type.IsObject():
			rp.Value, err = parseObject(param, header, ",", param.IsExplode(), defaultDecode)
		default:
			rp.Value, err = p.ParseWith(header, schema.ConvertToJsonSchema(param.Schema))
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}
