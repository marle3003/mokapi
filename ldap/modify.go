package ldap

import (
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type Operation int64

const (
	AddOperation     Operation = 0
	DeleteOperation  Operation = 1
	ReplaceOperation Operation = 2
)

type ModifyRequest struct {
	Dn    string             `json:"dn"`
	Items []ModificationItem `json:"changes"`
}

type ModificationItem struct {
	Operation    Operation    `json:"operation"`
	Modification Modification `json:"modification"`
}

type Modification struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

type ModifyResponse struct {
	ResultCode uint8  `json:"resultCode"`
	MatchedDn  string `json:"matchedDn"`
	Message    string `json:"message"`
}

func decodeModifyRequest(body *ber.Packet) (*ModifyRequest, error) {
	r := &ModifyRequest{}

	if len(body.Children) != 2 {
		return nil, fmt.Errorf("unexpected modify request length: %v", len(body.Children))
	}

	r.Dn = body.Children[0].Value.(string)
	mods := body.Children[1]

	for _, c := range mods.Children {
		item := ModificationItem{
			Operation: Operation(c.Children[0].Value.(int64)),
		}
		m := c.Children[1]
		mod := Modification{
			Type: m.Children[0].Value.(string),
		}
		for _, v := range m.Children[1].Children {
			mod.Values = append(mod.Values, v.Value.(string))
		}
		item.Modification = mod
		r.Items = append(r.Items, item)
	}

	return r, nil
}

func (r *ModifyRequest) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, modifyRequest, nil, "Modify Request")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Dn, "object"))

	items := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "items")
	for _, c := range r.Items {
		item := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "item")
		item.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(c.Operation), "operation"))
		item.AppendChild(c.Modification.encode())

		items.AppendChild(item)
	}

	body.AppendChild(items)

	envelope.AppendChild(body)
}

func (m *Modification) encode() *ber.Packet {
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "modification")
	seq.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, m.Type, "modification"))

	values := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "values")
	for _, v := range m.Values {
		values.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v, "value"))
	}
	seq.AppendChild(values)
	return seq
}

func decodeModifyResponse(p *ber.Packet) (*ModifyResponse, error) {
	code := p.Children[0].Value.(int64)

	r := &ModifyResponse{
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

func (r *ModifyResponse) encode(envelope *ber.Packet) {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, modifyResponse, nil, "Modify Response")
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(r.ResultCode), "ResultCode"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.MatchedDn, "MatchedDN"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Message, "DiagnosticMessage"))
	envelope.AppendChild(body)
}
