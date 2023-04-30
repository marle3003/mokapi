package ldap

import ber "gopkg.in/go-asn1-ber/asn1-ber.v1"

type UnbindRequest struct {
}

func (r *UnbindRequest) toPacket() *ber.Packet {
	return ber.Encode(ber.ClassApplication, ber.TypePrimitive, unbindRequest, nil, "Unbind Request")
}
