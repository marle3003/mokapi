package ldap

import ber "gopkg.in/go-asn1-ber/asn1-ber.v1"

type CompareRequest struct {
	Dn        string `json:"dn"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type CompareResponse struct {
	ResultCode uint8  `json:"resultCode"`
	Message    string `josn:"message"`
}

func decodeCompareRequest(body *ber.Packet) (*CompareRequest, error) {
	r := &CompareRequest{
		Dn:        body.Children[0].Value.(string),
		Attribute: body.Children[1].Value.(string),
		Value:     body.Children[2].Value.(string),
	}
	return r, nil
}

func (r *CompareRequest) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, compareRequest, nil, "compare request")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Dn, "DN"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Attribute, "attribute"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Value, "value"))
	envelope.AppendChild(body)
}

func decodeCompareResponse(p *ber.Packet) (*CompareResponse, error) {
	code := p.Children[0].Value.(int64)

	r := &CompareResponse{
		ResultCode: uint8(code),
	}
	if len(p.Children) > 1 {
		r.Message = p.Children[1].Value.(string)
	}

	return r, nil
}

func (r *CompareResponse) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, compareResponse, nil, "Compare Response")
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(r.ResultCode), "ResultCode"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Message, "DiagnosticMessage"))
	envelope.AppendChild(body)
}
