package mqtt

type DisconnectRequest struct {
	Reason     uint8
	Properties Properties
}

func (r *DisconnectRequest) Read(d *Decoder, _ *Header) {
	if d.leftSize > 0 {
		r.Reason = d.ReadByte()
	}
	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}
}

func (r *DisconnectRequest) Write(e *Encoder, _ *Header) {
	if r.Reason != 0 {
		e.writeByte(r.Reason)
	}
}
