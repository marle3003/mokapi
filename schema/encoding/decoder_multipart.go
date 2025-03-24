package encoding

import (
	"bytes"
	"io"
	"mime/multipart"
	"mokapi/media"
)

type DecodePart func(m map[string]interface{}, part *multipart.Part) error

type MultipartDecoder struct {
}

func (d *MultipartDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.Type == "multipart"
}

func (d *MultipartDecoder) Decode(b []byte, state *DecodeState) (interface{}, error) {
	boundary := state.contentType.Parameters["boundary"]
	reader := multipart.NewReader(bytes.NewReader(b), boundary)

	m := make(map[string]interface{})
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if state.decodePart != nil {
			err = state.decodePart(m, part)
			if err != nil {
				return nil, err
			}
		} else {
			var v interface{}
			v, err = d.decodePart(part, state)
			if err != nil {
				return nil, err
			}
			m[part.FormName()] = v
		}

		part.Close()
	}
	return state.parser.Parse(m)
}

func (d *MultipartDecoder) decodePart(part *multipart.Part, state *DecodeState) (interface{}, error) {
	defer part.Close()
	b, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	ct := media.ParseContentType(part.Header.Get("Content-Type"))

	return Decode(b, withState(state), WithContentType(ct))
}
