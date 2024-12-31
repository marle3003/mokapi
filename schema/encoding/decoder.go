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

type DecodeFunc func(propName string, val interface{}) (interface{}, error)

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
			if err != nil {
				return nil, err
			}
			return state.parser.Parse(v)
		}
	}
	return state.parser.Parse(string(b))
}

func WithContentType(contentType media.ContentType) DecodeOptions {
	return func(state *DecodeState) {
		state.contentType = contentType
	}
}

func WithDecodeProperty(decodeFunc DecodeFunc) DecodeOptions {
	return func(state *DecodeState) {
		state.decodeProperty = decodeFunc
	}
}

func WithParser(p Parser) DecodeOptions {
	return func(state *DecodeState) {
		state.parser = p
	}
}

type DecodeState struct {
	contentType    media.ContentType
	decodeProperty DecodeFunc
	parser         Parser
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
	d.decodeProperty = nil
	d.parser = &parser.Parser{}
}
