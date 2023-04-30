package ldap

import ber "gopkg.in/go-asn1-ber/asn1-ber.v1"

type AbandonRequest struct {
	MessageId int64
}

func (r *AbandonRequest) toPacket() *ber.Packet {
	p := ber.Encode(ber.ClassApplication, ber.TypePrimitive, abandonRequest, nil, "Abandon Search")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.MessageId, "Message ID"))
	return p
}
