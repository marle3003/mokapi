package mqtt

import (
	"mokapi/buffer"
)

type MessageWriter interface {
	Write(e *Encoder)
}

type Response struct {
	Header  *Header
	Message any
}

type response struct {
	h       *Header
	session *ClientSession
}

func (r *response) Write(messageType Type, msg MessageWriter) {
	b := buffer.NewPageBuffer()

	e := NewEncoder(b)
	msg.Write(e)

	r.h.Type = messageType
	r.h.Size = b.Size()

	r.session.sendOrQueue(&packet{
		header:  r.h.with(messageType, b.Size()),
		payload: b,
	})
}
