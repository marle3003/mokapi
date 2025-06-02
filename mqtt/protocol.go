package mqtt

import "io"

type Type byte

const (
	CONNECT     Type = 1
	CONNACK     Type = 2
	PUBLISH     Type = 3
	PUBACK      Type = 4
	PUBREC      Type = 5
	PUBREL      Type = 6
	PUBCOMP     Type = 7
	SUBSCRIBE   Type = 8
	SUBACK      Type = 9
	UNSUBSCRIBE Type = 10
	UNSUBACK    Type = 11
	PINGREQ     Type = 12
	PINGRESP    Type = 13
	DISCONNECT  Type = 14
)

type Header struct {
	Type   Type
	Size   int
	Dup    bool
	Qos    byte
	Retain bool
}

func readHeader(d *Decoder) *Header {
	h := &Header{}

	b := d.ReadByte()
	h.Type = Type(b >> 4)
	h.Dup = (b>>3)&0x01 > 0
	h.Qos = (b >> 0x06) >> 1
	h.Retain = b&0x01 != 0
	h.Size = d.readRemainingLength()

	d.leftSize = h.Size

	return h
}

func (h *Header) Write(w io.Writer) error {
	b := byte(h.Type)<<4 | encodeBool(h.Dup)<<3 | h.Qos<<1 | encodeBool(h.Retain)
	w.Write([]byte{b})
	return writeRemainingLength(w, h.Size)
}

func (h *Header) with(messageType Type, size int) *Header {
	return &Header{
		Type:   messageType,
		Size:   size,
		Dup:    h.Dup,
		Qos:    h.Qos,
		Retain: h.Retain,
	}
}

func ReadResponse(r io.Reader) *Response {
	d := NewDecoder(r, 5)
	res := &Response{
		Header: readHeader(d),
	}

	switch res.Header.Type {
	case CONNACK:
		c := &ConnectResponse{}
		c.Read(d)
		res.Message = c
	case SUBACK:
		c := &SubscribeResponse{}
		c.Read(d)
		res.Message = c
	}

	return res
}

type Code struct {
	Reason string
	Code   byte
}

var (
	Accepted                      = Code{"accepted", 0x0}
	ErrUnsupportedProtocolVersion = Code{"unacceptable protocol version", 0x1}
	ErrIdentifierRejected         = Code{"identifier rejected", 0x2}
)
