package ldap

import ber "gopkg.in/go-asn1-ber/asn1-ber.v1"

type ModifyDNRequest struct {
	Dn            string
	NewRdn        string
	DeleteOldDn   bool
	NewSuperiorDn string
}

type ModifyDNResponse struct {
	ResultCode uint8
	MatchedDn  string
	Message    string
}

func decodeModifyDNRequest(body *ber.Packet) (*ModifyDNRequest, error) {
	r := &ModifyDNRequest{
		Dn:          body.Children[0].Value.(string),
		NewRdn:      body.Children[1].Value.(string),
		DeleteOldDn: body.Children[2].Value.(bool),
	}
	if len(body.Children) >= 4 {
		r.NewSuperiorDn = body.Children[3].Value.(string)
	}
	return r, nil
}

func (m *ModifyDNRequest) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, modifyDNRequest, nil, "modify DN")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, m.Dn, "DN"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, m.NewRdn, "new Rdn"))
	if m.DeleteOldDn {
		body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, string([]byte{0xff}), "delete old DN"))
	} else {
		body.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, m.DeleteOldDn, "delete old DN"))
	}
	if m.NewSuperiorDn != "" {
		body.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, ber.TagEOC, m.NewSuperiorDn, "new Superior DN"))
	}
	envelope.AppendChild(body)
}

func decodeModifyDnResponse(p *ber.Packet) (*ModifyDNResponse, error) {
	code := p.Children[0].Value.(int64)

	r := &ModifyDNResponse{
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

func (r *ModifyDNResponse) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, modifyDNResponse, nil, "Modify DN Response")
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(r.ResultCode), "ResultCode"))
	envelope.AppendChild(body)
}
