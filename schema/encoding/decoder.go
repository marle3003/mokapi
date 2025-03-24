package encoding

import (
	"io"
	"mokapi/media"
	"mokapi/schema/json/parser"
	"sync"
)

type Decoder interface {
	IsSupporting(contentType media.ContentType) bool
	Decode([]byte, *DecodeState) (interface{}, error)
}

type Parser interface {
	Parse(data interface{}) (interface{}, error)
}

var decoders []Decoder

func init() {
	RegisterDecoder(&JsonDecoder{})
	RegisterDecoder(&FormUrlEncodeDecoder{})
	RegisterDecoder(&MultipartDecoder{})
	RegisterDecoder(&BinaryDecoder{})
	RegisterDecoder(&TextDecoder{})
	RegisterDecoder(&AvroDecoder{})
}

func RegisterDecoder(d Decoder) {
	decoders = append(decoders, d)
}

type DecodeOptions func(state *DecodeState)

func DecodeFrom(r io.Reader, opts ...DecodeOptions) (interface{}, error) {
	if r == nil {
		return Decode(nil, opts...)
	}

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
			v, err := d.Decode(b, state)
			return v, err
		}
	}
	return state.parser.Parse(b)
}

func WithContentType(contentType media.ContentType) DecodeOptions {
	return func(state *DecodeState) {
		state.contentType = contentType
	}
}

func WithDecodePart(decodeFunc DecodePart) DecodeOptions {
	return func(state *DecodeState) {
		state.decodePart = decodeFunc
	}
}

func WithDecodeFormUrlParam(decodeFunc DecodeFormUrlParam) DecodeOptions {
	return func(state *DecodeState) {
		state.decodeFormUrlParam = decodeFunc
	}
}

func WithParser(p Parser) DecodeOptions {
	return func(state *DecodeState) {
		state.parser = p
	}
}

func withState(s *DecodeState) DecodeOptions {
	return func(state *DecodeState) {
		*state = *s
	}
}

type DecodeState struct {
	contentType        media.ContentType
	decodePart         DecodePart
	decodeFormUrlParam DecodeFormUrlParam
	parser             Parser
}

var decodeStatePool sync.Pool

func newDecodeState() *DecodeState {
	if v := decodeStatePool.Get(); v != nil {
		d := v.(*DecodeState)
		d.Reset()
		return d
	}
	d := &DecodeState{}
	d.Reset()
	return d
}

func (d *DecodeState) Reset() {
	d.contentType = media.Empty
	d.decodePart = nil
	d.parser = &parser.Parser{}
}
