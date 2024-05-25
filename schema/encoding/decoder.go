package encoding

import (
	"io"
	"mokapi/media"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"sync"
)

type Decoder interface {
	IsSupporting(contentType media.ContentType) bool
	Decode([]byte, media.ContentType, DecodeFunc) (interface{}, error)
}

type DecodeFunc func(propName string, val interface{}) (interface{}, error)

var decoders []Decoder

func init() {
	RegisterDecoder(&JsonDecoder{})
	RegisterDecoder(&FormUrlEncodeDecoder{})
	RegisterDecoder(&MultipartDecoder{})
	RegisterDecoder(&FileDecoder{})
	RegisterDecoder(&TextDecoder{})
}

func RegisterDecoder(d Decoder) {
	decoders = append(decoders, d)
}

type DecodeOptions func(state *decodeState)

func DecodeFrom(r io.Reader, opts ...DecodeOptions) (interface{}, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Decode(b, opts...)
}

func DecodeString(s string, opts ...DecodeOptions) (interface{}, error) {
	return Decode([]byte(s), opts...)
}

func Decode(b []byte, opts ...DecodeOptions) (interface{}, error) {
	state := newDecodeState()
	for _, opt := range opts {
		opt(state)
	}

	for _, d := range decoders {
		if d.IsSupporting(state.contentType) {
			v, err := d.Decode(b, state.contentType, state.decodeProperty)
			if err != nil {
				return nil, err
			}
			return state.parser.Parse(v, state.schema)
		}
	}
	return state.parser.Parse(string(b), state.schema)
}

func WithContentType(contentType media.ContentType) DecodeOptions {
	return func(state *decodeState) {
		state.contentType = contentType
	}
}

func WithDecodeProperty(decodeFunc DecodeFunc) DecodeOptions {
	return func(state *decodeState) {
		state.decodeProperty = decodeFunc
	}
}

func WithSchema(s *schema.Ref) DecodeOptions {
	return func(state *decodeState) {
		state.schema = s
	}
}

func WithParser(p *parser.Parser) DecodeOptions {
	return func(state *decodeState) {
		state.parser = p
	}
}

type decodeState struct {
	schema         *schema.Ref
	contentType    media.ContentType
	decodeProperty DecodeFunc
	parser         *parser.Parser
}

var decodeStatePool sync.Pool

func newDecodeState() *decodeState {
	if v := decodeStatePool.Get(); v != nil {
		d := v.(*decodeState)
		d.Reset()
		return d
	}
	d := &decodeState{}
	d.Reset()
	return d
}

func (d *decodeState) Reset() {
	d.schema = nil
	d.contentType = media.Empty
	d.decodeProperty = nil
	d.parser = &parser.Parser{}
}
