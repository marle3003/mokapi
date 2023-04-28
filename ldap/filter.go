package ldap

import (
	"bytes"
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"strings"
)

func compileFilter(filter string) (*ber.Packet, int, error) {
	if len(filter) == 0 || filter[0] != '(' {
		return nil, 0, fmt.Errorf("filter syntax error: expected starting with ( got %v", filter)
	}

	var v *bytes.Buffer
	var p *ber.Packet
	for pos := 0; pos < len(filter); pos++ {
		c := rune(filter[pos])
		switch {
		case c == '(':
			v = bytes.NewBuffer(nil)
		case c == ')':
			if v != nil {
				s := v.String()
				if p.Tag == ber.Tag(FilterEqualityMatch) && strings.Contains(s, "*") {
					p = compileStarFilter(p.Children[0].Value.(string), s)
				} else {
					p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, s, "Condition"))
				}
			}
			return p, pos + 1, nil
		case c == '=':
			p = ber.Encode(ber.ClassContext, ber.TypeConstructed, FilterEqualityMatch, nil, "Equality Match")
			p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v.String(), "Attribute"))
			v = bytes.NewBuffer(nil)
		case c == '<' && filter[pos+1] == '=':
			pos++
			p = ber.Encode(ber.ClassContext, ber.TypeConstructed, FilterLessOrEqual, nil, "Equality Match")
			p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v.String(), "Attribute"))
			v = bytes.NewBuffer(nil)
		case c == '!':
			p = ber.Encode(ber.ClassContext, ber.TypeConstructed, FilterNot, nil, "Filter Not")
			child, n, err := compileFilter(filter[pos+1:])
			if err != nil {
				return nil, 0, err
			}
			p.AppendChild(child)
			return p, pos + n + 2, nil
		case c == '&':
			p = ber.Encode(ber.ClassContext, ber.TypeConstructed, FilterAnd, nil, "Filter And")
			n, err := compileFilterSet(filter[pos+1:], p)
			return p, pos + n + 2, err
		case c == '|':
			p = ber.Encode(ber.ClassContext, ber.TypeConstructed, FilterOr, nil, "Filter Or")
			n, err := compileFilterSet(filter[pos+1:], p)
			return p, pos + n + 2, err
		default:
			v.WriteRune(c)
		}
	}
	return nil, 0, fmt.Errorf("unexpected filter end: %v", filter)
}

func compileFilterSet(filter string, p *ber.Packet) (int, error) {
	pos := 0
	for pos < len(filter) && filter[pos] != ')' {
		child, n, err := compileFilter(filter[pos:])

		if err != nil {
			return 0, err
		}
		p.AppendChild(child)
		pos += n
	}
	return pos, nil
}

func compileStarFilter(attr, value string) *ber.Packet {
	if value == "*" {
		return ber.NewString(ber.ClassContext, ber.TypePrimitive, FilterPresent, attr, "Filter present")
	}

	p := ber.Encode(ber.ClassContext, ber.TypeConstructed, FilterSubstrings, nil, "Substring")
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, attr, "Attribute"))
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Substrings")
	parts := strings.Split(value, "*")
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		tag := FilterSubstringsAny
		switch i {
		case 0:
			tag = FilterSubstringsStartWith
		case len(parts) - 1:
			tag = FilterSubstringsEndWith
		}
		seq.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, ber.Tag(tag), part, SubstringText[tag]))
	}
	p.AppendChild(seq)
	return p
}

func decompileFilter(p *ber.Packet) (string, error) {
	switch p.Tag {
	case FilterAnd:
		s, err := parseMultary(p.Children)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(&%v)", s), nil
	case FilterOr:
		s, err := parseMultary(p.Children)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(|%v)", s), nil
	case FilterNot:
		s, err := parseUnary(p.Children)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(!%v)", s), nil
	case FilterEqualityMatch:
		if len(p.Children) != 2 {
			return "", fmt.Errorf("invalid filter operation")
		}
		return fmt.Sprintf("(%v=%v)",
			p.Children[0].Value.(string),
			p.Children[1].Value.(string)), nil
	case FilterGreaterOrEqual:
		return fmt.Sprintf("(%v>=%v)",
			p.Children[0].Value.(string),
			p.Children[1].Value.(string)), nil
	case FilterLessOrEqual:
		return fmt.Sprintf("(%v<=%v)",
			p.Children[0].Value.(string),
			p.Children[1].Value.(string)), nil
	case FilterPresent:
		return fmt.Sprintf("(%v=*)", p.Data.String()), nil
	case FilterSubstrings:
		var sb strings.Builder
		for i, part := range p.Children[1].Children {
			b := part.Data.Bytes()
			val := string(b)
			switch uint8(part.Tag) {
			case FilterSubstringsStartWith:
				sb.WriteString(fmt.Sprintf("%v*", val))
			case FilterSubstringsEndWith:
				if i > 0 {
					sb.WriteString(val)
				} else {
					sb.WriteString(fmt.Sprintf("*%v", val))
				}
			default:
				format := "*%v*"
				if i > 0 {
					format = "%v*"
				}
				sb.WriteString(fmt.Sprintf(format, val))
			}
		}
		return fmt.Sprintf("(%v=%v)", p.Children[0].Value.(string), sb.String()), nil
	default:
		return "", fmt.Errorf("unsupported filter %v requested", p.Tag)
	}
}

func parseMultary(children []*ber.Packet) (string, error) {
	var sb strings.Builder
	for _, child := range children {
		s, err := decompileFilter(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	return sb.String(), nil
}

func parseUnary(children []*ber.Packet) (string, error) {
	if len(children) != 1 {
		return "", fmt.Errorf("invalid filter operation")
	}
	return decompileFilter(children[0])
}

func is(p *ber.Packet, op int) bool {
	return p != nil && p.Tag == ber.Tag(op)
}
