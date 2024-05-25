package encoding

import (
	"bytes"
	"io"
	"mime/multipart"
	"mokapi/media"
)

type MultipartDecoder struct {
}

func (d *MultipartDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.Type == "multipart"
}

func (d *MultipartDecoder) Decode(b []byte, contentType media.ContentType, decode DecodeFunc) (i interface{}, err error) {
	boundary := contentType.Parameters["boundary"]
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

		v, err := d.decodePart(part, decode)
		if err != nil {
			return nil, err
		}
		m[part.FormName()] = v
	}
	return m, nil
}

func (d *MultipartDecoder) decodePart(part *multipart.Part, decode DecodeFunc) (interface{}, error) {
	defer part.Close()
	b, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	if decode != nil {
		name := part.FormName()
		return decode(name, b)
	}

	ct := media.ParseContentType(part.Header.Get("Content-Type"))
	return Decode(b, WithContentType(ct), WithDecodeProperty(decode))
}
