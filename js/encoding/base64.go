package encoding

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"github.com/dop251/goja"
	"mokapi/engine/common"
)

type base64 struct {
	host common.Host
	rt   *goja.Runtime
}

func (b *base64) Encode(input interface{}) string {
	data, err := toBytes(input)
	if err != nil {
		panic(b.rt.ToValue(fmt.Errorf("base64 encode failed: %v", err)))
	}
	return b64.StdEncoding.EncodeToString(data)
}

func (b *base64) Decode(s string) string {
	d, err := b64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(b.rt.ToValue(err.Error()))
	}
	return string(d)
}

func toBytes(input interface{}) ([]byte, error) {
	switch v := input.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case goja.ArrayBuffer:
		return v.Bytes(), nil
	case []interface{}:
		var buf bytes.Buffer
		for _, v := range v {
			b, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("input type is not []byte")
			}
			buf.WriteByte(byte(b))
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("type not supported: %T", v)
	}
}
