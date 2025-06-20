package kafka

import (
	"encoding/binary"
	"fmt"
	"io"
	"mokapi/buffer"
)

type Response struct {
	Header  *Header
	Message Message
}

func NewResponse(key ApiKey, version int16, correlationId int32) *Response {
	return &Response{Header: &Header{
		ApiKey:        key,
		ApiVersion:    version,
		CorrelationId: correlationId,
	}}
}

func (r *Response) Read(reader io.Reader) error {
	d := NewDecoder(reader, 4)

	if r.Header == nil {
		return fmt.Errorf("header not set")
	}

	r.Header.Size = d.ReadInt32()
	if r.Header.Size == 0 {
		return io.EOF
	}
	d.leftSize = int(r.Header.Size) - 4

	correlationId := d.ReadInt32()
	if correlationId != r.Header.CorrelationId {
		return fmt.Errorf("error correlation id, expected %v, got %v, requested %v", r.Header.CorrelationId, correlationId, r.Header.ApiKey)
	}

	if r.Header.ApiVersion >= ApiTypes[r.Header.ApiKey].flexibleResponse {
		r.Header.TagFields = d.ReadTagFields()
	}

	if r.Header.Size == 0 {
		return io.EOF
	}

	if d.err != nil {
		return d.err
	}

	apiType := ApiTypes[r.Header.ApiKey]
	var err error
	r.Message, err = apiType.response.decode(d, r.Header.ApiVersion)
	if err != nil {
		return err
	}
	return d.err
}

func (r *Response) Write(w io.Writer) error {
	b := buffer.NewPageBuffer()
	defer b.Unref()

	e := NewEncoder(b)
	apiType := ApiTypes[r.Header.ApiKey]

	e.writeInt32(0) // placeholder length
	e.writeInt32(r.Header.CorrelationId)
	if r.Header.ApiVersion >= apiType.flexibleResponse {
		e.writeUVarInt(0) // tag_buffer
	}

	if r.Message != nil {
		if err := apiType.response.encode(e, r.Header.ApiVersion, r.Message); err != nil {
			return err
		}
	}

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(b.Size()-4))
	b.WriteAt(size[:], 0)

	_, err := b.WriteTo(w)

	return err
}
