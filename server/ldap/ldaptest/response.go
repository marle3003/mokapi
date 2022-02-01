package ldaptest

import ber "gopkg.in/go-asn1-ber/asn1-ber.v1"

type ResponseRecorder struct {
	Responses []Response
}

type Response struct {
	MessageId int64
	Body      *ber.Packet
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

func (r *ResponseRecorder) Write(packet *ber.Packet) error {
	r.Responses = append(r.Responses, Response{
		MessageId: packet.Children[0].Value.(int64),
		Body:      packet.Children[1],
	})
	return nil
}
