package mqtt

type UnsubscribeRequest struct {
	MessageId int16
	Topics    []string
}

func (r *UnsubscribeRequest) Read(d *Decoder) {
	r.MessageId = d.ReadInt16()

	for d.leftSize > 0 {
		name := d.ReadString()
		r.Topics = append(r.Topics, name)
	}
}

func (r *UnsubscribeRequest) Write(e *Encoder) {
	e.writeInt16(r.MessageId)
	for _, name := range r.Topics {
		e.writeString(name)
	}
}

type UnsubscribeResponse struct {
	MessageId int16
}

func (r *UnsubscribeResponse) Write(e *Encoder) {
	e.writeInt16(r.MessageId)
}

func (r *UnsubscribeResponse) Read(d *Decoder) {
	r.MessageId = d.ReadInt16()
}
