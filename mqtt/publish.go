package mqtt

type PublishRequest struct {
	Topic     string
	MessageId int16
	Data      []byte
}

func (r *PublishRequest) Read(d *Decoder) {
	r.Topic = d.ReadString()
	r.MessageId = d.ReadInt16()
	r.Data = make([]byte, d.leftSize)
	d.readFull(r.Data)
}

func (r *PublishRequest) Write(e *Encoder) {
	e.writeString(r.Topic)
	e.writeInt16(r.MessageId)
	e.Write(r.Data)
}

type PublishResponse struct {
	MessageId int16
}

func (r *PublishResponse) Read(d *Decoder) {
	r.MessageId = d.ReadInt16()
}

func (r *PublishResponse) Write(e *Encoder) {
	e.writeInt16(r.MessageId)
}
