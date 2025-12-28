package ldap

import (
	"fmt"

	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

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
	r := &ModifyDNRequest{}

	// 1. Entry (DN)
	entry, ok := body.Children[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("modifyDN request: entry is not a string")
	}
	r.Dn = entry

	// 2. New RDN
	newRDN, ok := body.Children[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("modifyDN request: newRDN is not a string")
	}
	r.NewRdn = newRDN

	// 3. deleteOldRDN
	deleteOld, ok := body.Children[2].Value.(bool)
	if !ok {
		// Some BER implementations decode BOOLEAN as uint8
		if b, ok2 := body.Children[2].Value.(uint8); ok2 {
			r.DeleteOldDn = b != 0
		} else {
			return nil, fmt.Errorf("modifyDN request: deleteOldRDN not a bool")
		}
	} else {
		r.DeleteOldDn = deleteOld
	}

	// 4. Optional newSuperior
	if len(body.Children) == 4 {
		// newSuperior is tagged with context-specific [0]
		ns := body.Children[3]
		if ns.Tag != 0 || ns.ClassType != ber.ClassContext {
			return nil, fmt.Errorf("modifyDN request: invalid newSuperior tag")
		}

		newSup, ok := ns.Value.(string)
		if !ok {
			newSup = string(ns.Data.Bytes())
		}

		r.NewSuperiorDn = newSup
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
