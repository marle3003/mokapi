package ldap

import (
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type AddRequest struct {
	Dn         string
	Attributes []Attribute
}

type Attribute struct {
	Type   string
	Values []string
}

type AddResponse struct {
	ResultCode uint8
	MatchedDn  string
	Message    string
}

func decodeAddRequest(body *ber.Packet) (*AddRequest, error) {
	r := &AddRequest{
		Dn: body.Children[0].Value.(string),
	}
	for _, attr := range body.Children[1].Children {
		a := &Attribute{
			Type: attr.Children[0].Value.(string),
		}
		for _, v := range attr.Children[1].Children {
			a.Values = append(a.Values, v.Value.(string))
		}
		r.Attributes = append(r.Attributes, *a)
	}
	return r, nil
}

func (r *AddRequest) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, addRequest, nil, "add request")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Dn, "DN"))
	attrs := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "attributes")
	for _, attr := range r.Attributes {
		attrs.AppendChild(attr.encode())
	}
	body.AppendChild(attrs)
	envelope.AppendChild(body)
}

func (m *Attribute) encode() *ber.Packet {
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "attribute")
	seq.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, m.Type, "type"))
	values := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "attribute values")
	for _, v := range m.Values {
		values.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v, "value"))
	}
	seq.AppendChild(values)
	return seq
}

func decodeAddResponse(p *ber.Packet) (*AddResponse, error) {
	code := p.Children[0].Value.(int64)

	r := &AddResponse{
		ResultCode: uint8(code),
	}
	if len(p.Children) > 1 {
		r.MatchedDn = p.Children[1].Value.(string)
	}
	if len(p.Children) > 2 {
		r.Message = p.Children[2].Value.(string)
	}

	return r, nil
}

func (r *AddResponse) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, addResponse, nil, "Add Response")
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(r.ResultCode), "ResultCode"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.MatchedDn, "MatchedDN"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Message, "DiagnosticMessage"))
	envelope.AppendChild(body)
}
