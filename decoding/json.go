package decoding

import (
	"encoding/json"
	"fmt"
	"mokapi/media"
)

type JsonDecoder struct {
}

func (d *JsonDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.Subtype == "json"
}

func (d *JsonDecoder) Decode(b []byte, _ media.ContentType, _ DecodeFunc) (i interface{}, err error) {
	err = json.Unmarshal(b, &i)
	if err != nil {
		err = fmt.Errorf("invalid json format: %v", err)
	}
	return
}
