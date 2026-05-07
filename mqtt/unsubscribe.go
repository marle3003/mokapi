package mqtt

type UnsubscriptionReason byte

const (
	UnsubscribeSuccess    UnsubscriptionReason = 0
	NoSubscriptionExisted UnsubscriptionReason = 17
)

type UnsubscribeRequest struct {
	MessageId  uint16
	Topics     []string
	Properties Properties
}

func (r *UnsubscribeRequest) Read(d *Decoder, _ *Header) {
	r.MessageId = d.ReadUInt16()

	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}
	for d.leftSize > 0 {
		name := d.ReadString()
		r.Topics = append(r.Topics, name)
	}
}

func (r *UnsubscribeRequest) Write(e *Encoder, _ *Header) {
	e.writeUInt16(r.MessageId)

	if e.IsV5() {
		r.Properties.Write(e)
	}
	for _, name := range r.Topics {
		e.writeString(name)
	}
}

type UnsubscribeResponse struct {
	MessageId   uint16
	ReasonCodes []UnsubscriptionReason
	Properties  Properties
}

func (r *UnsubscribeResponse) Write(e *Encoder, _ *Header) {
	e.writeUInt16(r.MessageId)

	if e.IsV5() {
		r.Properties.Write(e)

		for _, reason := range r.ReasonCodes {
			e.writeByte(byte(reason))
		}
	}
}

func (r *UnsubscribeResponse) Read(d *Decoder, _ *Header) {
	r.MessageId = d.ReadUInt16()

	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}

	for d.leftSize > 0 {
		code := d.ReadByte()
		r.ReasonCodes = append(r.ReasonCodes, UnsubscriptionReason(code))
	}
}
