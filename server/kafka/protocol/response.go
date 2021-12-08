package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
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

	r.Header.Size = d.readInt32()
	if r.Header.Size == 0 {
		return nil
	}
	d.leftSize = int(r.Header.Size)

	correlationId := d.readInt32()
	if correlationId != r.Header.CorrelationId {
		return fmt.Errorf("error correlation id, expected %v, got %v, requested %v", r.Header.CorrelationId, correlationId, r.Header.ApiKey)
	}

	if r.Header.ApiVersion >= ApiTypes[r.Header.ApiKey].flexibleResponse {
		r.Header.TagFields = d.readTagFields()
	}

	if r.Header.Size == 0 {
		return io.EOF
	}

	if d.err != nil {
		return d.err
	}

	apiType := ApiTypes[r.Header.ApiKey]
	r.Message = apiType.response.decode(d, r.Header.ApiVersion)

	return d.err
}

func (r *Response) Write(w io.Writer) error {
	buffer := newPageBuffer()
	defer func() {
		buffer.unref()
	}()

	e := NewEncoder(buffer)
	apiType := ApiTypes[r.Header.ApiKey]

	e.writeInt32(0) // placeholder length
	e.writeInt32(r.Header.CorrelationId)
	if r.Header.ApiVersion >= apiType.flexibleResponse {
		e.writeUVarInt(0) // tag_buffer
	}
	apiType.response.encode(e, r.Header.ApiVersion, r.Message)

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(buffer.Size()-4))
	buffer.WriteAt(size[:], 0)

	_, err := buffer.WriteTo(w)

	return err
}
