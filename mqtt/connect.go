package mqtt

type ConnectRequest struct {
	Header   ConnectHeader
	ClientId string
	Topic    string
	Message  string
	Username string
	Password string
}

type ConnectHeader struct {
	Protocol     string
	Version      byte
	Username     bool
	Password     bool
	WillRetain   bool
	WillQoS      byte
	WillFlag     bool
	CleanSession bool
	KeepAlive    int16
}

func readConnect(d *Decoder) *ConnectRequest {
	r := &ConnectRequest{}

	r.Header.Protocol = d.ReadString()
	r.Header.Version = d.ReadByte()

	b := d.ReadByte()
	r.Header.Username = (b>>7)&0x1 > 0
	r.Header.Password = (b>>6)&0x1 > 0
	r.Header.WillRetain = (b>>5)&0x1 > 0
	r.Header.WillQoS = (b >> 3) & 0x3
	r.Header.WillFlag = (b>>2)&0x1 > 0
	r.Header.CleanSession = (b>>1)&0x1 > 0
	r.Header.KeepAlive = d.ReadInt16()

	r.ClientId = d.ReadString()

	if r.Header.WillFlag {
		r.Topic = d.ReadString()
		r.Message = d.ReadString()
	}

	if r.Header.Username {
		r.Username = d.ReadString()
	}
	if r.Header.Password {
		r.Password = d.ReadString()
	}

	return r
}

func (r *ConnectRequest) Write(e *Encoder) {
	e.writeString(r.Header.Protocol)
	e.writeByte(r.Header.Version)
	b := byte(0)
	if r.Header.Username {
		b |= 0x1 << 7
	}
	if r.Header.Password {
		b |= 0x1 << 6
	}
	if r.Header.WillRetain {
		b |= 0x1 << 5
	}
	b |= (r.Header.WillQoS & 0x03) << 1
	if r.Header.WillFlag {
		b |= 0x1 << 2
	}
	if r.Header.CleanSession {
		b |= 0x1 << 1
	}
	e.writeByte(b)
	e.writeInt16(r.Header.KeepAlive)
	e.writeString(r.ClientId)

	if r.Header.Username {
		e.writeString(r.Username)
	}
	if r.Header.Password {
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
