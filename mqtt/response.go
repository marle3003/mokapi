package mqtt

import (
	"mokapi/buffer"
)

type Response struct {
	Header  *Header
	Message Message
}

type response struct {
	h   *Header
	ctx *ClientContext
}

func (r *response) Write(messageType Type, msg Message) {
	b := buffer.NewPageBuffer()

	e := NewEncoder(b)
	msg.Write(e)

	r.h.Type = messageType
	r.h.Size = b.Size()

	r.ctx.sendOrQueue(&packet{
		header:  r.h.with(messageType, b.Size()),
		payload: b,
	})
}
