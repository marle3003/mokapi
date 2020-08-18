package ldap

import (
	"fmt"

	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

func parseSearchRequest(req *ber.Packet) (*SearchRequest, error) {
	if len(req.Children) != 8 {
		return nil, fmt.Errorf("Unexpected search request length: %v", len(req.Children))
	}

	baseObject, ok := req.Children[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("Unexpected data type for field baseobject: %v", req.Children[0].Value)
	}

	s, ok := req.Children[1].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("Unexpected data type for field scope: %v", req.Children[1].Value)
	}
	scope := int(s)

	d, ok := req.Children[2].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("Unexpected data type for field dereference policy: %v", req.Children[2].Value)
	}
	derefPolicy := int(d)

	s, ok = req.Children[3].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("Unexpected data type for field size limit: %v", req.Children[3].Value)
	}
	sizeLimit := int(s)

	t, ok := req.Children[4].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("Unexpected data type for field time limit: %v", req.Children[4].Value)
	}
	timeLimit := int(t)

	typesOnly := false
	if req.Children[5].Value != nil {
		typesOnly, ok = req.Children[5].Value.(bool)
		if !ok {
			return nil, fmt.Errorf("Unexpected data type for field types: %v", req.Children[4].Value)
		}
	}

	filterPacket := req.Children[6]
	filter, _ := parseFilter(filterPacket)

	attributes := []string{}
	for _, attr := range req.Children[7].Children {
		a, ok := attr.Value.(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected data type for field attributes: %v", req.Children[4].Value)
		}
		attributes = append(attributes, a)
	}

	return &SearchRequest{
		baseObject:        baseObject,
		scope:             scope,
		dereferencePolicy: derefPolicy,
		sizeLimit:         sizeLimit,
		timeLimit:         timeLimit,
		typesOnly:         typesOnly,
		filter:            filterPacket,
		filterString:      filter,
		attributes:        attributes,
	}, nil
}

func parseFilter(packet *ber.Packet) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ldap: error decompiling filter")
		}
	}()
	ret = "("
	err = nil
	childStr := ""

	switch packet.Tag {
	case FilterAnd:
		ret += "&"
		for _, child := range packet.Children {
			childStr, err = parseFilter(child)
			if err != nil {
				return
			}
			ret += childStr
		}
	case FilterOr:
		ret += "|"
		for _, child := range packet.Children {
			childStr, err = parseFilter(child)
			if err != nil {
				return
			}
			ret += childStr
		}
	case FilterNot:
		ret += "!"
		childStr, err = parseFilter(packet.Children[0])
		if err != nil {
			return
		}
		ret += childStr

	case FilterSubstrings:
		ret += ber.DecodeString(packet.Children[0].Data.Bytes())
		ret += "="
		switch packet.Children[1].Children[0].Tag {
		case FilterSubstringsStartWith:
			ret += ber.DecodeString(packet.Children[1].Children[0].Data.Bytes()) + "*"
		case FilterSubstringsAny:
			ret += "*" + ber.DecodeString(packet.Children[1].Children[0].Data.Bytes()) + "*"
		case FilterSubstringsEndWith:
			ret += "*" + ber.DecodeString(packet.Children[1].Children[0].Data.Bytes())
		}
	case FilterEqualityMatch:
		ret += ber.DecodeString(packet.Children[0].Data.Bytes())
		ret += "="
		ret += ber.DecodeString(packet.Children[1].Data.Bytes())
	case FilterGreaterOrEqual:
		ret += ber.DecodeString(packet.Children[0].Data.Bytes())
		ret += ">="
		ret += ber.DecodeString(packet.Children[1].Data.Bytes())
	case FilterLessOrEqual:
		ret += ber.DecodeString(packet.Children[0].Data.Bytes())
		ret += "<="
		ret += ber.DecodeString(packet.Children[1].Data.Bytes())
	case FilterPresent:
		ret += ber.DecodeString(packet.Data.Bytes())
		ret += "=*"
	case FilterApproxMatch:
		ret += ber.DecodeString(packet.Children[0].Data.Bytes())
		ret += "~="
		ret += ber.DecodeString(packet.Children[1].Data.Bytes())
	}

	ret += ")"
	return
}
