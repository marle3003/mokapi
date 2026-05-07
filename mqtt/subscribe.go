package mqtt

type SubscriptionReason byte

const (
	GrantedQoS0      SubscriptionReason = 0
	GrantedQoS1      SubscriptionReason = 1
	GrantedQoS2      SubscriptionReason = 2
	UnspecifiedError SubscriptionReason = 128
)

type SubscribeRequest struct {
	MessageId  uint16
	Topics     []SubscribeTopic
	Properties Properties
}

type SubscribeTopic struct {
	Name string
	QoS  byte
}

func (r *SubscribeRequest) Read(d *Decoder, _ *Header) {
	r.MessageId = d.ReadUInt16()

	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}

	for d.leftSize > 0 {
		name := d.ReadString()
		qos := d.ReadByte()
		r.Topics = append(r.Topics, SubscribeTopic{
			Name: name,
			QoS:  qos,
		})
	}
}

func (r *SubscribeRequest) Write(e *Encoder, _ *Header) {
	e.writeUInt16(r.MessageId)

	if e.IsV5() {
		r.Properties.Write(e)
	}

	for _, t := range r.Topics {
		e.writeString(t.Name)
		e.writeByte(t.QoS)
	}
}

type SubscribeResponse struct {
	MessageId   uint16
	ReasonCodes []SubscriptionReason
	Properties  Properties
}

func (r *SubscribeResponse) Write(e *Encoder, _ *Header) {
	e.writeUInt16(r.MessageId)

	if e.IsV5() {
		r.Properties.Write(e)
	}

	for _, reason := range r.ReasonCodes {
		e.writeByte(byte(reason))
	}
}

func (r *SubscribeResponse) Read(d *Decoder, _ *Header) {
	r.MessageId = d.ReadUInt16()

	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}

	for d.leftSize > 0 {
		code := d.ReadByte()
		r.ReasonCodes = append(r.ReasonCodes, SubscriptionReason(code))
	}
}
