package ldap

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type Control interface {
	ControlType() string
	Encode() *ber.Packet
}

const (
	PagedResultsControlType = "1.2.840.113556.1.4.319"
)

type PagedResultsControl struct {
	Criticality bool
	PageSize    int64
	Cookie      string
}

func decodeControls(p *ber.Packet) ([]Control, error) {
	var controls []Control

	for _, c := range p.Children {
		if len(c.Children) == 0 {
			continue
		}
		oid, ok := c.Children[0].Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid control type: expected OID string but got %T", c.Children[0].Value)
		}
		switch oid {
		case PagedResultsControlType:
			ctrl := &PagedResultsControl{}
			err := ctrl.Decode(c)
			if err != nil {
				return nil, err
			}
			controls = append(controls, ctrl)
		default:
			log.Errorf("LDAP control '%v' not supported", oid)
		}
	}

	return controls, nil
}

func (p *PagedResultsControl) ControlType() string {
	return PagedResultsControlType
}

func (p *PagedResultsControl) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "1.2.840.113556.1.4.319", "Control Type (Paging)"))
	if p.Criticality {
		packet.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, p.Criticality, "Criticality"))
	}

	p2 := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Control Value (Paging)")
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Search Control Value")
	seq.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, p.PageSize, "Paging Size"))
	cookie := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Cookie")
	cookie.Value = p.Cookie
	cookie.Data.Write([]byte(p.Cookie))
	seq.AppendChild(cookie)
	p2.AppendChild(seq)

	packet.AppendChild(p2)

	return packet
}

func (p *PagedResultsControl) Decode(packet *ber.Packet) error {
	if len(packet.Children) < 2 {
		return fmt.Errorf("invalid control type: expected at least 2 children but got %d", len(packet.Children))
	}

	valueIndex := 1
	if len(packet.Children) > 2 {
		// criticality is optional
		p.Criticality = packet.Children[1].Value.(bool)
		valueIndex = 2
	}

	var value *ber.Packet
	if packet.Children[valueIndex].Tag == ber.TagSequence {
		value = packet.Children[valueIndex]
	} else {
		// try to decode from data
		// this case happens using VSCode extension LDAP Explorer
		value = ber.DecodePacket(packet.Children[valueIndex].Data.Bytes())
	}

	if len(value.Children) != 2 {
		return fmt.Errorf("expected 2 children (INTEGER, OCTET STRING), got %v", len(packet.Children))
	}

	var ok bool
	p.PageSize, ok = value.Children[0].Value.(int64)
	if !ok {
		return fmt.Errorf("expected int64 for page size, got %T", value.Children[0].Value)
	}
	p.Cookie, ok = value.Children[1].Value.(string)
	if !ok {
		return fmt.Errorf("expected string for cookie, got %T", value.Children[1].Value)
	}

	return nil
}

const (
	pagingContextKey = "PagingContext"
)

type PagingContext struct {
	Cookies map[string]int64
}

func PagingFromContext(ctx context.Context) *PagingContext {
	return ctx.Value(pagingContextKey).(*PagingContext)
}

func NewPagingFromContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, pagingContextKey, &PagingContext{Cookies: make(map[string]int64)})
}
