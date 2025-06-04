package mqtt

import (
	"context"
	"fmt"
	"io"
	"mokapi/buffer"
)

type Request struct {
	Header  *Header
	Message Message
	Context context.Context
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Context = ctx
	return r
}

func (r *Request) Write(w io.Writer) error {
	b := buffer.NewPageBuffer()
	defer b.Unref()

	e := NewEncoder(b)
	r.Message.Write(e)

	r.Header.Size = b.Size()
	err := r.Header.Write(w)
	if err != nil {
		return err
	}

	_, err = b.WriteTo(w)
	return err
}

func (r *Request) Read(reader io.Reader) error {
	d := NewDecoder(reader, 5)
	r.Header = readHeader(d)
	if d.err != nil {
		return d.err
	}
	d.leftSize = r.Header.Size

	switch r.Header.Type {
	case CONNECT:
		connect := &ConnectRequest{}
		connect.Read(d)
		r.Message = connect
		client := ClientFromContext(r.Context)
		if client != nil {
			client.ClientId = connect.ClientId
		}
	case PUBLISH:
		publish := &PublishRequest{}
		publish.Read(d)
		r.Message = publish
	case SUBSCRIBE:
		subscribe := &SubscribeRequest{}
		subscribe.Read(d)
		r.Message = subscribe
	case UNSUBSCRIBE:
		unsubscribe := &UnsubscribeRequest{}
		unsubscribe.Read(d)
		r.Message = unsubscribe
	}

	if d.err != nil {
		return d.err
	}

	if d.leftSize > 0 {
		return fmt.Errorf("mqtt: remaining length %d is not zero", d.leftSize)
	}

	return nil
}
