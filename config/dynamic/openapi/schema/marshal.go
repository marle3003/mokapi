package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/media"
	"mokapi/sortedmap"
)

type marshalObject struct {
	*sortedmap.LinkedHashMap[string, interface{}]
}

func (r *Ref) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	i, err := selectData(i, r)
	if err != nil {
		return nil, fmt.Errorf("serialize data to '%v' failed: %w", contentType.String(), err)
	}
	switch {
	case contentType.Subtype == "json" || contentType.Subtype == "problem+json":
		b, err := json.Marshal(i)
		if err, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("json error (%v): %v", err.Offset, err.Error())
		}
		return b, err
	case contentType.IsXml():
		return writeXml(i, r)
	default:
		if s, ok := i.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}
}

func (m *schemaObject) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')
	l := m.Len()
	i := 0
	for it := m.Iter(); it.Next(); {
		k := it.Key()
		v := it.Value()

		s := fmt.Sprintf("%v", k)

		if s == "minimum" {
			fmt.Sprintf("")
		}

		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteRune(':')
		value, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buf.Write(value)
		if i != l-1 {
			buf.WriteRune(',')
		}
		i++
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}
