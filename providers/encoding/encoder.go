package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
)

type Encoder interface {
	Encode(i interface{}, schema *openapi.SchemaRef) ([]byte, error)
}

func Encode(i interface{}, contentType *media.ContentType, schema *openapi.SchemaRef) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		b, err := MarshalJSON(i, schema)
		if err, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("json error (%v): %v", err.Offset, err.Error())
		}
		return b, err
	case "xml", "rss+xml":
		var buffer bytes.Buffer
		w := newXmlWriter(&buffer)
		err := w.write(i, schema)
		if err != nil {
			return nil, err
		}
		return buffer.Bytes(), nil
	default:
		if s, ok := i.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}

	return nil, fmt.Errorf("unsupported content type %v", contentType)
}

func toString(i interface{}) string {
	switch v := i.(type) {
	case float64:
		if i := math.Trunc(v); i == v {
			return fmt.Sprintf("%v", int64(i))
		}
		return fmt.Sprintf("%f", v)
	case float32:
		f := float64(v)
		if i := math.Trunc(f); i == f {
			return fmt.Sprintf("%v", int64(i))
		}
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf("%v", i)
	}
}
