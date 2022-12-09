package ldap

import (
	"fmt"
	ldapConfig "mokapi/config/dynamic/ldap"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type SearchRequest struct {
	BaseDN            string
	Scope             int64
	DereferencePolicy int64
	SizeLimit         int64
	TimeLimit         int64
	TypesOnly         bool
	Filter            *ber.Packet
	Attributes        []string
}

type SearchResult struct {
	dn         string
	attributes map[string][]string
}

type predicate func(entry ldapConfig.Entry) bool

func (d *Directory) search(rw ResponseWriter, r *Request) error {
	searchRequest, err := parseSearchRequest(r.Body)
	if err != nil {
		return rw.Write(newSearchDone(r.MessageId, ProtocolError, err.Error()))
	}

	filterString, err := parseFilter(searchRequest.Filter)
	if err != nil {
		log.Debugf("unable to parse filter: %v", err)
	}
	log.Infof("ldap search request: messageId=%v BaseDN=%v Filter=%v",
		r.MessageId, searchRequest.BaseDN, filterString)
	d.monitor.Search.WithLabel(d.config.Info.Name).Add(1)
	d.monitor.LastSearch.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))

	if searchRequest.BaseDN == "" && searchRequest.Scope == ScopeBaseObject {
		result := &SearchResult{dn: "", attributes: make(map[string][]string)}
		for k, v := range d.config.Root.Attributes {
			result.attributes[k] = v
		}

		searchResult := newSearchResult(r.MessageId, result)
		if err := rw.Write(searchResult); err != nil {
			return err
		}
		log.Infof("found result for message %v: %v", r.MessageId, result.dn)
		return rw.Write(newSearchDone(r.MessageId, ResultSuccess, ""))
	}

	count := int64(0)
	predicate, err := filter(searchRequest.Filter)
	if err != nil {
		return err
	}
	for _, entry := range d.config.Entries {
		if !predicate(entry) {
			continue
		}

		switch searchRequest.Scope {
		case ScopeBaseObject:
			if entry.Dn != searchRequest.BaseDN {
				continue
			}
		case ScopeSingleLevel:
			parts := strings.Split(entry.Dn, ",")
			if len(parts) < 2 && entry.Dn != searchRequest.BaseDN {
				continue
			}
			if dn := strings.Join(parts[1:], ","); dn != searchRequest.BaseDN {
				continue
			}
		}

		if count > searchRequest.SizeLimit {
			return nil
		}

		result := &SearchResult{dn: entry.Dn, attributes: make(map[string][]string)}

		result.attributes["objectClass"] = entry.Attributes["objectClass"]

		// todo
		if len(searchRequest.Attributes) > 0 {
			for _, a := range searchRequest.Attributes {
				for k, v := range entry.Attributes {
					if strings.ToLower(a) == strings.ToLower(k) {
						result.attributes[a] = v
					}
				}
			}
		} else {
			result.attributes = entry.Attributes
		}

		log.Infof("found result for message %v: %v", r.MessageId, result.dn)

		response := newSearchResult(r.MessageId, result)
		if err := rw.Write(response); err != nil {
			return err
		}
	}

	return rw.Write(newSearchDone(r.MessageId, ResultSuccess, ""))
}

func newSearchDone(messageID int64, status int64, msg string) *ber.Packet {
	responsePacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	responsePacket.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "Message ID"))
	donePacket := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationSearchResultDone, nil, "Search result done")
	donePacket.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, status, "resultCode: "))
	donePacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matchedDN: "))
	donePacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, msg, "errorMessage: "))
	responsePacket.AppendChild(donePacket)

	return responsePacket
}

func newSearchResult(messageId int64, result *SearchResult) *ber.Packet {
	responsePacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	responsePacket.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageId, "Message ID"))

	searchEntry := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationSearchResultEntry, nil, "Search Result Entry")
	searchEntry.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, result.dn, "Object Name"))

	attrs := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attributes:")
	for k, v := range result.attributes {
		if k == "dn" {
			continue
		}
		attrs.AppendChild(newSearchAttribute(k, v))
	}

	searchEntry.AppendChild(attrs)
	responsePacket.AppendChild(searchEntry)

	return responsePacket
}

func newSearchAttribute(name string, values []string) *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attribute")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, name, "Attribute Name"))

	valuesPacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "Attribute Values")
	for _, value := range values {
		valuesPacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, value, "Attribute Value"))
	}

	packet.AppendChild(valuesPacket)

	return packet
}

func filter(f *ber.Packet) (predicate, error) {
	switch f.Tag {
	case FilterAnd:
		var predicates []predicate
		for _, child := range f.Children {
			p, err := filter(child)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, p)
		}
		return func(entry ldapConfig.Entry) bool {
			for _, p := range predicates {
				if !p(entry) {
					return false
				}
			}
			return true
		}, nil
	case FilterOr:
		var predicates []predicate
		for _, child := range f.Children {
			p, err := filter(child)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, p)
		}
		return func(entry ldapConfig.Entry) bool {
			for _, p := range predicates {
				if p(entry) {
					return true
				}
			}
			return false
		}, nil
	case FilterNot:
		if len(f.Children) != 1 {
			return nil, fmt.Errorf("invalid filter operation")
		}

		p, err := filter(f.Children[0])
		if err != nil {
			return nil, err
		}
		return func(entry ldapConfig.Entry) bool {
			return !p(entry)
		}, nil
	case FilterEqualityMatch:
		attribute := f.Children[0].Value.(string)
		value := f.Children[1].Value.(string)

		return func(entry ldapConfig.Entry) bool {
			for k, list := range entry.Attributes {
				if strings.ToLower(attribute) == strings.ToLower(k) {
					for _, v := range list {
						if strings.ToLower(value) == strings.ToLower(v) {
							return true
						}
					}
				}
			}
			return false
		}, nil

	case FilterSubstrings:
		if len(f.Children) != 2 {
			return nil, fmt.Errorf("invalid filter operation")
		}
		attribute := f.Children[0].Value.(string)

		var predicates []func(string) bool
		for _, cond := range f.Children[1].Children {
			bytes := cond.Data.Bytes()
			value := string(bytes[:])
			var p func(string, string) bool
			switch cond.Tag {
			case FilterSubstringsStartWith:
				p = strings.HasPrefix
			case FilterSubstringsAny:
				p = strings.Contains
			case FilterSubstringsEndWith:
				p = strings.HasSuffix
			default:
				return nil, fmt.Errorf("unsupported substring %v", cond.Tag)
			}
			predicates = append(predicates, func(s string) bool {
				return p(s, value)
			})
		}

		return func(entry ldapConfig.Entry) bool {
			for k, list := range entry.Attributes {
				if strings.ToLower(attribute) == strings.ToLower(k) {
					for _, v := range list {
						for _, p := range predicates {
							if !p(v) {
								return false
							}
						}
						return true
					}
				}
			}
			return false
		}, nil
	case FilterGreaterOrEqual:
		return nil, fmt.Errorf("not supported")
	case FilterLessOrEqual:
		return nil, fmt.Errorf("not supported")
	case FilterPresent:
		attribute := strings.ToLower(f.Data.String())
		return func(entry ldapConfig.Entry) bool {
			for k := range entry.Attributes {
				if strings.ToLower(k) == attribute {
					return true
				}
			}
			return false
		}, nil
	}

	return nil, fmt.Errorf("unsupported filter %v requested", f.Tag)
}

func parseSearchRequest(req *ber.Packet) (*SearchRequest, error) {
	if len(req.Children) != 8 {
		return nil, fmt.Errorf("unexpected search request length: %v", len(req.Children))
	}

	baseDN, ok := req.Children[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field baseobject: %v", req.Children[0].Value)
	}

	scope, ok := req.Children[1].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field scope: %v", req.Children[1].Value)
	}

	derefPolicy, ok := req.Children[2].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field dereference policy: %v", req.Children[2].Value)
	}

	sizeLimit, ok := req.Children[3].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field size limit: %v", req.Children[3].Value)
	}

	timeLimit, ok := req.Children[4].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for field time limit: %v", req.Children[4].Value)
	}

	typesOnly := false
	if req.Children[5].Value != nil {
		typesOnly, ok = req.Children[5].Value.(bool)
		if !ok {
			return nil, fmt.Errorf("unexpected data type for field types only: %v", req.Children[4].Value)
		}
	}

	var attributes []string
	for _, attr := range req.Children[7].Children {
		a, ok := attr.Value.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected data type for field attributes: %v", req.Children[4].Value)
		}
		attributes = append(attributes, a)
	}

	return &SearchRequest{
		BaseDN:            baseDN,
		Scope:             scope,
		DereferencePolicy: derefPolicy,
		SizeLimit:         sizeLimit,
		TimeLimit:         timeLimit,
		TypesOnly:         typesOnly,
		Filter:            req.Children[6],
		Attributes:        attributes,
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
