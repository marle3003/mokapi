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
	QoS    byte
	Retain bool
}

func readHeader(d *Decoder) *Header {
	h := &Header{}

	b := d.ReadByte()
	h.Type = Type(b >> 4)
	h.Dup = (b>>3)&0x01 > 0
	h.QoS = (b >> 0x06) >> 1
	h.Retain = b&0x01 != 0
	h.Size = d.readRemainingLength()

	d.leftSize = h.Size

	return h
}

func (h *Header) Write(w io.Writer) error {
	b := byte(h.Type)<<4 | encodeBool(h.Dup)<<3 | h.QoS<<1 | encodeBool(h.Retain)
	w.Write([]byte{b})
	return writeRemainingLength(w, h.Size)
}

func (h *Header) with(messageType Type, size int) *Header {
	return &Header{
		Type:   messageType,
		Size:   size,
		Dup:    h.Dup,
		QoS:    h.QoS,
		Retain: h.Retain,
	}
}

type Code struct {
	Reason string
	Code   byte
}

var (
	Accepted                      = Code{Code: 0x00, Reason: "accepted"}
	ErrUnsupportedProtocolVersion = Code{Code: 0x01, Reason: "unacceptable protocol version"}
	ErrIdentifierRejected         = Code{Code: 0x02, Reason: "identifier rejected"}
	ErrUnspecifiedError           = Code{Code: 0x80, Reason: "unspecified error"}
)
