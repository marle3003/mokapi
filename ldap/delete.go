package ldap

import (
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type DeleteRequest struct {
	Dn string `json:"dn"`
}

type DeleteResponse struct {
	ResultCode uint8  `json:"resultCode"`
	MatchedDn  string `json:"matchedDn"`
	Message    string `json:"message"`
}

func decodeDeleteRequest(body *ber.Packet) (*DeleteRequest, error) {
	r := &DeleteRequest{
		Dn: body.Children[0].Value.(string),
	}
	return r, nil
}

func (r *DeleteRequest) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, deleteRequest, nil, "delete request")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Dn, "DN"))
	envelope.AppendChild(body)
}

func decodeDeleteResponse(p *ber.Packet) (*DeleteResponse, error) {
	code := p.Children[0].Value.(int64)

	r := &DeleteResponse{
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

func (r *DeleteResponse) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, deleteResponse, nil, "Delete Response")
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(r.ResultCode), "ResultCode"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.MatchedDn, "MatchedDN"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Message, "DiagnosticMessage"))
	envelope.AppendChild(body)
}
