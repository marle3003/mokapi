package openapi

import (
	"fmt"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"net/http"
)

func parseQueryString(param *Parameter, r *http.Request) (*RequestParameterValue, error) {
	q := r.URL.RawQuery
	if len(param.Content) != 1 {
		return nil, fmt.Errorf("querystring parameter must contain exactly one content")
	}
	var mt *MediaType
	for _, v := range param.Content {
		mt = v
		break
	}
	if mt == nil {
		return nil, fmt.Errorf("missing media type for querystring parameter")
	}

	p := parser.Parser{Schema: schema.ConvertToJsonSchema(mt.Schema), ValidateAdditionalProperties: true}
	opts := []encoding.DecodeOptions{
		encoding.WithContentType(mt.ContentType),
		encoding.WithParser(&p),
	}

	if mt.ContentType.Type == "text" {
		p.ConvertStringToNumber = true
	}
	if mt.ContentType.String() == "application/x-www-form-urlencoded" {
		p.ConvertStringToNumber = true
		opts = append(opts, encoding.WithDecodeFormUrlParam(urlValueDecoder{mt: mt}.decode))
	}

	v, err := encoding.Decode([]byte(q), opts...)
	if err != nil {
		return nil, err
	}
	return &RequestParameterValue{
		Value: v,
		Raw:   &q,
	}, nil
}
