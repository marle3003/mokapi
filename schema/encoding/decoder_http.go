package encoding

import (
	"mokapi/media"
	"net/url"
)

type DecodeFormUrlParam func(name string, value interface{}) (interface{}, error)

type FormUrlEncodeDecoder struct {
}

func (d *FormUrlEncodeDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.String() == "application/x-www-form-urlencoded"
}

func (d *FormUrlEncodeDecoder) Decode(b []byte, state *DecodeState) (interface{}, error) {
	values, err := url.ParseQuery(string(b))
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	for k, v := range values {
		if state.decodeFormUrlParam != nil {
			i, err := state.decodeFormUrlParam(k, v)
			if err != nil {
				return nil, err
			}
			m[k] = i
		} else {
			m[k] = v
		}

	}
	return state.parser.Parse(m)
}
