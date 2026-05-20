package encoding

import (
	"encoding/json"
	"fmt"
	"mokapi/media"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"strconv"
	"strings"
)

type Encoder struct {
	r *schema.Schema
}

func NewEncoder(r *schema.Schema) *Encoder {
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
		return nil, err
	}

	var b []byte
	switch {
	case contentType.Subtype == "json" || strings.HasSuffix(contentType.Subtype, "+json"):
		b, err = json.Marshal(i)
	case contentType.Subtype == "xml" || strings.HasSuffix(contentType.Subtype, "+xml"):
		b, err = MarshalXml(i, e.r)
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
		return nil, err
	}
	return b, nil
}
