package mqtt

type DisconnectReason uint8

const (
	DisconnectNormal          DisconnectReason = iota
	DisconnectWithWillMessage DisconnectReason = 4
)

type DisconnectRequest struct {
	Reason     DisconnectReason
	Properties Properties
}

func (r *DisconnectRequest) Read(d *Decoder, _ *Header) {
	if d.leftSize > 0 {
		r.Reason = DisconnectReason(d.ReadByte())
	}
	if d.IsV5() {
		r.Properties = Properties{}
		r.Properties.Read(d)
	}
}

func (r *DisconnectRequest) Write(e *Encoder, _ *Header) {
	if r.Reason != 0 {
		e.writeByte(uint8(r.Reason))
	}
}
