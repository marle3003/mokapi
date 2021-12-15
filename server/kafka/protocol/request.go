package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type Request struct {
	Header  *Header
	Message Message
	Context Context
}

func (r *Request) Write(w io.Writer) error {
	if r.Message == nil {
		return fmt.Errorf("message is nil")
	}

	buffer := newPageBuffer()
	defer func() {
		buffer.unref()
	}()

	e := NewEncoder(buffer)
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
	binary.BigEndian.PutUint32(size[:], uint32(buffer.Size()-4))
	buffer.WriteAt(size[:], 0)

	_, err := buffer.WriteTo(w)

	if err != nil {
		return fmt.Errorf("kafka: Write apikey %v: %v", r.Header.ApiKey, err)
	}

	return nil
}
