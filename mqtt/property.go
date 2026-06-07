package mqtt

import "bytes"

const (
	PayloadFormatIndicator byte = 0x1
	ContentType            byte = 0x3
	SessionExpiryInterval  byte = 0x11
	ReasonString           byte = 0x1F
	UserProperty           byte = 0x26
)

type Properties map[byte]any

func (p Properties) Read(d *Decoder) {
	propLen := d.ReadVariableInt()
	if propLen == 0 {
		return
	}

	stopAt := d.leftSize - propLen
	for d.leftSize > stopAt && d.leftSize > 0 {
		propID := d.ReadByte()
		switch propID {
		case SessionExpiryInterval:
			p[propID] = d.ReadInt32()
		case ReasonString, ContentType:
			p[propID] = d.ReadString()
		case UserProperty:
			if p[propID] == nil {
				p[propID] = map[string]string{}
			}
			key := d.ReadString()
			val := d.ReadString()
			p[propID].(map[string]string)[key] = val
		}
	}
}

func (p Properties) Write(e *Encoder) {
	if len(p) == 0 {
		e.WriteVariableInt(0)
		return
	}

	var b bytes.Buffer
	propBuffer := NewEncoder(&b, e.protocolVersion)
	for id, val := range p {
		propBuffer.writeByte(id)

		switch v := val.(type) {
		case string:
			propBuffer.writeString(v)
		case int32:
			propBuffer.writeInt32(v)
		}
	}

	e.WriteVariableInt(b.Len())
	e.Write(b.Bytes())
}

func (p Properties) SessionExpiryInterval() int32 {
	if p == nil {
		return 0
	}
	v, ok := p[SessionExpiryInterval]
	if !ok {
		return 0
	}
	return v.(int32)
}

func (p Properties) ReasonString() string {
	if p == nil {
		return ""
	}
	v, ok := p[ReasonString]
	if !ok {
		return ""
	}
	return v.(string)
}
