package ldap

import (
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type SearchRequest struct {
	BaseDN            string    `json:"baseDN"`
	Scope             int64     `json:"scope"`
	DereferencePolicy int64     `json:"dereferencePolicy"`
	SizeLimit         int64     `json:"sizeLimit"`
	TimeLimit         int64     `json:"timeLimit"`
	TypesOnly         bool      `json:"typesOnly"`
	Filter            string    `json:"filter"`
	Attributes        []string  `json:"attributes"`
	Controls          []Control `json:"controls"`
}

type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Status   uint8          `json:"status"`
	Message  string         `json:"message"`
	Controls []Control      `json:"controls"`
}

type SearchResult struct {
	Dn         string              `json:"dn"`
	Attributes map[string][]string `json:"attributes"`
}

func NewSearchResult(dn string) SearchResult {
	return SearchResult{Dn: dn, Attributes: map[string][]string{}}
}

func decodeSearchRequest(p *ber.Packet, controls []Control) (*SearchRequest, error) {
	if len(p.Children) != 8 {
		return nil, fmt.Errorf("unexpected search request length: %v", len(p.Children))
	}

	baseDN, ok := p.Children[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field baseobject: %v", p.Children[0].Value)
	}

	scope, ok := p.Children[1].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field scope: %v", p.Children[1].Value)
	}

	derefPolicy, ok := p.Children[2].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field dereference policy: %v", p.Children[2].Value)
	}

	sizeLimit, ok := p.Children[3].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field size limit: %v", p.Children[3].Value)
	}

	timeLimit, ok := p.Children[4].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field time limit: %v", p.Children[4].Value)
	}

	typesOnly := false
	if p.Children[5].Value != nil {
		typesOnly, ok = p.Children[5].Value.(bool)
		if !ok {
			return nil, fmt.Errorf("unexpected data type for field types only: %v", p.Children[4].Value)
		}
	}

	var attributes []string
	for _, attr := range p.Children[7].Children {
		a, ok := attr.Value.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected data type for field attributes: %v", p.Children[4].Value)
		}
		attributes = append(attributes, a)
	}

	filter, err := decompileFilter(p.Children[6])
	if err != nil {
		return nil, err
	}

	return &SearchRequest{
		BaseDN:            baseDN,
		Scope:             scope,
		DereferencePolicy: derefPolicy,
		SizeLimit:         sizeLimit,
		TimeLimit:         timeLimit,
		TypesOnly:         typesOnly,
		Filter:            filter,
		Attributes:        attributes,
		Controls:          controls,
	}, nil
}

func (r *SearchRequest) encode(envelope *ber.Packet) error {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, searchRequest, nil, "Search Request")
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.BaseDN, "Base DN"))
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, r.Scope, "Scope"))
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, r.DereferencePolicy, "Dereference Policy"))
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.SizeLimit, "SizeLimit"))
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.TimeLimit, "TimeLimit"))
	body.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, r.TypesOnly, "TypesOnly"))

	f, _, err := compileFilter(r.Filter)
	if err != nil {
		return err
	}
	body.AppendChild(f)

	attributes := ber.NewSequence("Attributes")
	for _, attr := range r.Attributes {
		attributes.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, attr, "Attribute"))
	}
	body.AppendChild(attributes)
	envelope.AppendChild(body)

	if len(r.Controls) > 0 {
		controls := ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "Controls")

		for _, ctrl := range r.Controls {
			controls.AppendChild(ctrl.Encode())
		}

		envelope.AppendChild(controls)
	}

	return nil
}

func decodeSearchResponse(packets []*ber.Packet) (*SearchResponse, error) {
	res := &SearchResponse{}
	for _, p := range packets {
		if p.Tag == searchResult {
			r, err := decodeSearchResult(p)
			if err != nil {
				return nil, err
			}
			res.Results = append(res.Results, r)
		} else {
			res.Status = uint8(p.Children[0].Value.(int64))
			res.Message = p.Children[2].Value.(string)
		}

	}
	return res, nil
}

func decodeSearchResult(p *ber.Packet) (SearchResult, error) {
	r := NewSearchResult(p.Children[0].Value.(string))
	for _, c := range p.Children[1].Children {
		name := c.Children[0].Value.(string)
		var values []string
		for _, v := range c.Children[1].Children {
			values = append(values, v.Value.(string))
		}
		r.Attributes[name] = values
	}
	return r, nil
}

func encodeAttribute(name string, values []string) *ber.Packet {
	p := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attribute")
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, name, "Attribute Name"))

	valuesPacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "Attribute Values")
	for _, value := range values {
		valuesPacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, value, "Attribute Value"))
	}

	p.AppendChild(valuesPacket)

	return p
}

func (r *SearchResponse) appendSearchDone(envelope *ber.Packet) {
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, searchDone, nil, "Search result done")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, r.Status, "resultCode: "))
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matchedDN: "))
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Message, "errorMessage: "))
	envelope.AppendChild(p)

	if len(r.Controls) > 0 {
		controls := ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "Controls")

		for _, ctrl := range r.Controls {
			controls.AppendChild(ctrl.Encode())
		}

		envelope.AppendChild(controls)
	}

}

func (r *SearchResult) appendTo(envelope *ber.Packet) {
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, searchResult, nil, "Search Result Entry")
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Dn, "Object Name"))

	attrs := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attributes:")
	for k, v := range r.Attributes {
		if k == "dn" {
			continue
		}
		attrs.AppendChild(encodeAttribute(k, v))
	}

	p.AppendChild(attrs)
	envelope.AppendChild(p)
}
