package mqtt

type ConnectRequest struct {
	Protocol     string
	Version      byte
	HasUsername  bool
	HasPassword  bool
	WillRetain   bool
	WillQoS      byte
	WillFlag     bool
	CleanSession bool
	KeepAlive    int16
	ClientId     string
	Topic        string
	Message      string
	Username     string
	Password     string
}

type ConnectHeader struct {
}

func (r *ConnectRequest) Read(d *Decoder) {
	r.Protocol = d.ReadString()
	r.Version = d.ReadByte()

	b := d.ReadByte()
	r.HasUsername = (b>>7)&0x1 > 0
	r.HasPassword = (b>>6)&0x1 > 0
	r.WillRetain = (b>>5)&0x1 > 0
	r.WillQoS = (b >> 3) & 0x3
	r.WillFlag = (b>>2)&0x1 > 0
	r.CleanSession = (b>>1)&0x1 > 0
	r.KeepAlive = d.ReadInt16()

	r.ClientId = d.ReadString()

	if r.WillFlag {
		r.Topic = d.ReadString()
		r.Message = d.ReadString()
	}

	if r.HasUsername {
		r.Username = d.ReadString()
	}
	if r.HasPassword {
		r.Password = d.ReadString()
	}
}

func (r *ConnectRequest) Write(e *Encoder) {
	e.writeString(r.Protocol)
	e.writeByte(r.Version)
	b := byte(0)
	if r.HasUsername {
		b |= 0x1 << 7
	}
	if r.HasPassword {
		b |= 0x1 << 6
	}
	if r.WillRetain {
		b |= 0x1 << 5
	}
	b |= (r.WillQoS & 0x03) << 1
	if r.WillFlag {
		b |= 0x1 << 2
	}
	if r.CleanSession {
		b |= 0x1 << 1
	}
	e.writeByte(b)
	e.writeInt16(r.KeepAlive)
	e.writeString(r.ClientId)

	if r.HasUsername {
		e.writeString(r.Username)
	}
	if r.HasPassword {
		e.writeString(r.Password)
	}
}

type ConnectResponse struct {
	SessionPresent bool
	ReturnCode     Code
}

func (r *ConnectResponse) Write(e *Encoder) {
	if r.SessionPresent {
		e.writeByte(0x01)
	} else {
		e.writeByte(0x0)
	}
	e.writeByte(r.ReturnCode.Code)
}

func (r *ConnectResponse) Read(d *Decoder) {
	r.SessionPresent = d.ReadByte() == 0x01
	r.ReturnCode = Code{
		Code: d.ReadByte(),
	}
}
