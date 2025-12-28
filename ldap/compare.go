package ldap

import (
	"fmt"

	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

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
	if len(body.Children) != 2 {
		return nil, fmt.Errorf("invalid compare request: expected 2 children, got %d", len(body.Children))
	}

	dnPacket := body.Children[0]
	kvPacket := body.Children[1]

	if len(kvPacket.Children) != 2 {
		return nil, fmt.Errorf("invalid attribute value assertion: expected 2 children")
	}

	r := &CompareRequest{}

	// DN
	if dn, ok := dnPacket.Value.(string); ok {
		r.Dn = dn
	} else {
		return nil, fmt.Errorf("dn is not a string")
	}

	// Attribute
	if attr, ok := kvPacket.Children[0].Value.(string); ok {
		r.Attribute = attr
	} else {
		return nil, fmt.Errorf("attribute is not a string")
	}

	// Value (string or []byte depending on encoding)
	switch v := kvPacket.Children[1].Value.(type) {
	case string:
		r.Value = v
	case []byte:
		r.Value = string(v)
	default:
		return nil, fmt.Errorf("assertion value is not string or []byte")
	}

	return r, nil
}

func (r *CompareRequest) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, compareRequest, nil, "compare request")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Dn, "DN"))

	ava := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "AttributeValueAssertion")
	ava.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Attribute, "attribute"))
	ava.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Value, "value"))
	body.AppendChild(ava)

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
