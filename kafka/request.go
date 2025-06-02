package kafka

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"mokapi/buffer"
	"reflect"
)

type Request struct {
	Host    string
	Header  *Header
	Message Message
	Context context.Context
}

func (r *Request) Write(w io.Writer) error {
	if r.Message == nil {
		return fmt.Errorf("message is nil")
	}

	b := buffer.NewPageBuffer()
	defer b.Unref()

	e := NewEncoder(b)
	t := ApiTypes[r.Header.ApiKey]

	e.writeInt32(0) // placeholder length
	e.writeInt16(int16(r.Header.ApiKey))
	e.writeInt16(r.Header.ApiVersion)
	e.writeInt32(r.Header.CorrelationId)
	e.writeNullString(r.Header.ClientId)
	if r.Header.ApiVersion >= t.flexibleRequest {
		e.writeUVarInt(0) // tag_buffer
	}

	encode := newEncodeFunc(reflect.TypeOf(r.Message).Elem(), r.Header.ApiVersion, kafkaTag{})
	encode(e, reflect.ValueOf(r.Message).Elem())

	// update length
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(b.Size()-4))
	b.WriteAt(size[:], 0)

	_, err := b.WriteTo(w)

	if err != nil {
		return fmt.Errorf("kafka: Write apikey %v: %v", r.Header.ApiKey, err)
	}

	return nil
}

func (r *Request) Read(reader io.Reader) error {
	d := NewDecoder(reader, 4)
	r.Header = readHeader(d)

	if d.err != nil {
		return d.err
	}

	if r.Header.Size == 0 {
		return nil
	}

	if d.err != nil {
		return d.err
	}

	t := ApiTypes[r.Header.ApiKey]
	if t.MinVersion > r.Header.ApiVersion && t.MaxVersion < r.Header.ApiVersion {
		return Error{Header: r.Header, Code: UnsupportedVersion, Message: fmt.Sprintf("unsupported api version")}
	}

	var err error
	r.Message, err = t.request.decode(d, r.Header.ApiVersion)
	return err
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Context = ctx
	return r
}
