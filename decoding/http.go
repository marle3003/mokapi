package decoding

import (
	"mokapi/media"
	"net/url"
)

type FormUrlEncodeDecoder struct {
}

func (d *FormUrlEncodeDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.String() == "application/x-www-form-urlencoded"
}

func (d *FormUrlEncodeDecoder) Decode(b []byte, _ media.ContentType, decode DecodeFunc) (i interface{}, err error) {
	values, err := url.ParseQuery(string(b))
	m := map[string]interface{}{}
	for k, v := range values {
		if decode != nil {
			i, err = decode(k, v)
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
