package encoding

import (
	"bytes"
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
)

type Encoder interface {
	Encode(i interface{}, schema *openapi.SchemaRef) ([]byte, error)
}

func Encode(i interface{}, contentType *media.ContentType, schema *openapi.SchemaRef) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		return MarshalJSON(i, schema)
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
