package encoding

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

func (d *JsonDecoder) Decode(b []byte, state *DecodeState) (interface{}, error) {
	var i interface{}
	err := json.Unmarshal(b, &i)
	if err != nil {
		return nil, fmt.Errorf("invalid json format: %v", err)
	}
	return state.parser.Parse(i)
}
