package parameter

import (
	"fmt"
	"mokapi/config/dynamic/openapi/schema"
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
		switch p.Schema.Value.Type {
		case "array":
			rp.Value, err = parseArray(p, header, ",")
		case "object":
			rp.Value, err = parseObject(p, header, ",", p.IsExplode())
		default:
			rp.Value, err = schema.ParseString(header, p.Schema)
		}
	}

	if err != nil {
		return nil, err
	}

	return rp, nil
}
