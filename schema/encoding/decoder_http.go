package encoding

import (
	"mokapi/media"
	"net/url"
)

type FormUrlEncodeDecoder struct {
}

func (d *FormUrlEncodeDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.String() == "application/x-www-form-urlencoded"
}

func (d *FormUrlEncodeDecoder) Decode(b []byte, state *DecodeState) (i interface{}, err error) {
	values, err := url.ParseQuery(string(b))
	m := map[string]interface{}{}
	for k, v := range values {
		if state.decodeProperty != nil {
			i, err = state.decodeProperty(k, v)
			if err != nil {
				return nil, err
			}
			m[k] = i
		} else {
			m[k] = v
		}

	}
	return m, err
}
