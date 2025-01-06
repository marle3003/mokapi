package encoding

import (
	"encoding/json"
	"fmt"
	"mokapi/media"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"strconv"
)

const marshalError = "encoding data to '%v' failed: %w"
const marshalErrorList = "encoding data to '%v' failed:\n%w"

type Encoder struct {
	r *schema.Ref
}

func NewEncoder(r *schema.Ref) *Encoder {
	return &Encoder{
		r: r,
	}
}

func (e *Encoder) Write(v interface{}, contentType media.ContentType) ([]byte, error) {
	p := parser.Parser{ConvertToSortedMap: true}
	if contentType.Subtype != "json" {
		p.ConvertStringToNumber = true
	}

	i, err := p.ParseWith(v, e.r)
	if err != nil {
		if uw, ok := err.(interface{ Unwrap() []error }); ok {
			errs := uw.Unwrap()
			if len(errs) > 1 {
				return nil, fmt.Errorf(marshalErrorList, contentType.String(), err)
			}
		}

		return nil, fmt.Errorf(marshalError, contentType.String(), err)
	}
	var b []byte
	switch {
	case contentType.Subtype == "json" || contentType.Subtype == "problem+json":
		b, err = json.Marshal(i)
	default:
		var s string
		switch i.(type) {
		case string:
			s = i.(string)
		case float64:
			s = strconv.FormatFloat(i.(float64), 'f', -1, 64)
		case float32:
			s = strconv.FormatFloat(float64(i.(float32)), 'f', -1, 32)
		case int, int32, int64:
			s = fmt.Sprintf("%v", i)
		default:
			err = fmt.Errorf("not supported encoding of content types '%v', except simple data types", contentType)
		}
		b = []byte(s)
	}

	if err != nil {
		return nil, fmt.Errorf(marshalError, contentType.String(), err)
	}
	return b, nil
}
