package mqtt

type PublishReason byte

const (
	PublishSuccess       PublishReason = 0
	TopicNameInvalid     PublishReason = 144
	PayloadFormatInvalid PublishReason = 153
)

type PublishRequest struct {
	Topic      string
	MessageId  uint16
	Data       []byte
	Properties Properties
}

func (r *PublishRequest) Read(d *Decoder, h *Header) {
	r.Topic = d.ReadString()

	if h.QoS > 0 {
		r.MessageId = d.ReadUInt16()
	}

	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}

	r.Data = make([]byte, d.leftSize)
	d.readFull(r.Data)
}

func (r *PublishRequest) Write(e *Encoder, h *Header) {
	e.writeString(r.Topic)
	if h.QoS > 0 {
		e.writeUInt16(r.MessageId)
	}

	if e.IsV5() {
		r.Properties.Write(e)
	}

	e.Write(r.Data)
}

type PublishResponse struct {
	MessageId  uint16
	ReasonCode PublishReason
	Properties Properties
}

func (r *PublishResponse) Read(d *Decoder, h *Header) {
	r.MessageId = d.ReadUInt16()
	r.ReasonCode = PublishReason(d.ReadByte())

	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}
}

func (r *PublishResponse) Write(e *Encoder, _ *Header) {
	e.writeUInt16(r.MessageId)
	e.writeByte(byte(r.ReasonCode))
	if e.IsV5() {
		r.Properties.Write(e)
	}
}
