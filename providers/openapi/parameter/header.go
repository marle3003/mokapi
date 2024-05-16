package parameter

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"net/http"
)

func parseHeader(p *Parameter, r *http.Request) (*RequestParameterValue, error) {
	header := r.Header.Get(p.Name)

	if len(header) == 0 {
		if p.Required {
			return nil, fmt.Errorf("parameter is required")
		}
		return nil, nil
	}

	rp := &RequestParameterValue{Raw: header, Value: header}
	var err error
	if p.Schema != nil {
		switch {
		case p.Schema.Value.Type.IsArray():
			rp.Value, err = parseArray(p, header, ",")
		case p.Schema.Value.Type.IsObject():
			rp.Value, err = parseObject(p, header, ",", p.IsExplode(), defaultDecode)
		default:
			rp.Value, err = schema.ParseString(header, p.Schema)
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}
