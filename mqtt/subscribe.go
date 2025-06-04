package mqtt

type SubscribeRequest struct {
	MessageId int16
	Topics    []SubscribeTopic
}

type SubscribeTopic struct {
	Name string
	QoS  byte
}

func (r *SubscribeRequest) Read(d *Decoder) {
	r.MessageId = d.ReadInt16()

	for d.leftSize > 0 {
		name := d.ReadString()
		qos := d.ReadByte()
		r.Topics = append(r.Topics, SubscribeTopic{
			Name: name,
			QoS:  qos,
		})
	}
}

func (r *SubscribeRequest) Write(e *Encoder) {
	e.writeInt16(r.MessageId)
	for _, t := range r.Topics {
		e.writeString(t.Name)
		e.writeByte(t.QoS)
	}
}

type SubscribeResponse struct {
	MessageId int16
	TopicQoS  []byte
}

func (r *SubscribeResponse) Write(e *Encoder) {
	e.writeInt16(r.MessageId)
	for _, qos := range r.TopicQoS {
		e.writeByte(qos)
	}
}

func (r *SubscribeResponse) Read(d *Decoder) {
	r.MessageId = d.ReadInt16()
	for d.leftSize > 0 {
		r.TopicQoS = append(r.TopicQoS, d.ReadByte())
	}
}
