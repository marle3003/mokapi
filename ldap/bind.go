package ldap

import (
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type AuthType int

const (
	Simple AuthType = 0
)

type BindRequest struct {
	Version  int64
	Name     string
	Password string
	Auth     AuthType
}

type BindResponse struct {
	Result    uint8
	MatchedDN string
	Message   string
}

func readBindRequest(p *ber.Packet) (*BindRequest, error) {
	name, ok := p.Children[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("unable to parse name: expected string")
	}

	r := &BindRequest{
		Version:  p.Children[0].Value.(int64),
		Name:     name,
		Password: p.Children[2].Data.String(),
		Auth:     AuthType(p.Children[2].Tag),
	}
	return r, nil
}

func readBindResponse(p *ber.Packet) (*BindResponse, error) {
	return &BindResponse{
		Result:    uint8(p.Children[0].Value.(int64)),
		MatchedDN: p.Children[1].Value.(string),
		Message:   p.Children[2].Value.(string),
	}, nil
}

func (r *BindRequest) toPacket() *ber.Packet {
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, bindRequest, nil, "Bind Request")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.Version, "Version"))
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Name, "Username"))
	p.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, ber.Tag(r.Auth), r.Password, "Password"))
	return p
}

func (r *BindResponse) toPacket() *ber.Packet {
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, bindResponse, nil, "Bind Response")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, r.Result, "resultCode: "))
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.MatchedDN, "matchedDN: "))
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, r.Message, "errorMessage: "))
	return p
}
