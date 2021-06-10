package ldap

import (
	"fmt"
	ldapConfig "mokapi/config/dynamic/ldap"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type SearchRequest struct {
	baseObject        string
	scope             int
	dereferencePolicy int
	sizeLimit         int
	timeLimit         int
	typesOnly         bool
	filter            *ber.Packet
	filterString      string
	attributes        []string
}

type SearchResult struct {
	dn         string
	attributes map[string][]string
}

func (s *Binding) handleSearchRequest(conn net.Conn, messageId int64, req *ber.Packet) error {
	searchRequest, err := parseSearchRequest(req)
	if err != nil {
		return err
	}

	log.Infof("received search request with messageId %v BaseDN: %v Filter %v", messageId, searchRequest.baseObject, searchRequest.filterString)

	if searchRequest.baseObject == "" && searchRequest.scope == ScopeBaseObject {
		result := &SearchResult{dn: "", attributes: make(map[string][]string)}
		for k, v := range s.root.Attributes {
			result.attributes[k] = v
		}

		response := s.encodeSearchResult(messageId, result)
		sendResponse(conn, response)
		return nil
	}

	count := 0
	for _, entry := range s.entries {
		ok, err := filter(searchRequest.filter, entry)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		switch searchRequest.scope {
		case ScopeBaseObject:
			if entry.Dn != searchRequest.baseObject {
				continue
			}
		case ScopeSingleLevel:
			parts := strings.Split(entry.Dn, ",")
			if len(parts) < 2 && entry.Dn != searchRequest.baseObject {
				continue
			}
			if dn := strings.Join(parts[1:], ","); dn != searchRequest.baseObject {
				continue
			}
		}

		if count > searchRequest.sizeLimit {
			return nil
		}

		result := &SearchResult{dn: entry.Dn, attributes: make(map[string][]string)}

		result.attributes["objectClass"] = entry.Attributes["objectClass"]

		// todo
		if len(searchRequest.attributes) > 0 {
			for _, a := range searchRequest.attributes {
				for k, v := range entry.Attributes {
					if strings.ToLower(a) == strings.ToLower(k) {
						result.attributes[a] = v
					}
				}
			}
		} else {
			result.attributes = entry.Attributes
		}

		log.Infof("found result for message %v: %v", messageId, result.dn)

		response := s.encodeSearchResult(messageId, result)
		sendResponse(conn, response)
	}

	log.Debugf("finishing search for message %v", messageId)

	return nil
}

func filter(f *ber.Packet, entry ldapConfig.Entry) (bool, error) {
	switch f.Tag {
	case FilterAnd:
		for _, child := range f.Children {
			ok, err := filter(child, entry)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	case FilterOr:
		for _, child := range f.Children {
			ok, err := filter(child, entry)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	case FilterNot:
		if len(f.Children) != 1 {
			return false, fmt.Errorf("invalid filter operation")
		}

		ok, err := filter(f.Children[0], entry)
		if err != nil {
			return false, err
		}
		return !ok, nil
	case FilterEqualityMatch:
		attribute := f.Children[0].Value.(string)
		value := f.Children[1].Value.(string)

		for k, a := range entry.Attributes {
			if strings.ToLower(attribute) == strings.ToLower(k) {
				for _, v := range a {
					if strings.ToLower(value) == strings.ToLower(v) {
						return true, nil
					}
				}
			}
		}
	case FilterSubstrings:
		if len(f.Children) != 2 {
			return false, fmt.Errorf("invalid filter operation")
		}
		attribute := f.Children[0].Value.(string)
		bytes := f.Children[1].Children[0].Data.Bytes()
		value := string(bytes[:])
		for k, a := range entry.Attributes {
			if strings.ToLower(attribute) == strings.ToLower(k) {
				for _, v := range a {
					switch f.Children[1].Children[0].Tag {
					case FilterSubstringsStartWith:
						if strings.HasPrefix(v, value) {
							return true, nil
						}
					case FilterSubstringsAny:
						if strings.Contains(v, value) {
							return true, nil
						}
					case FilterSubstringsEndWith:
						if strings.HasSuffix(v, value) {
							return true, nil
						}
					}
				}
			}
		}
	case FilterGreaterOrEqual:
		return false, fmt.Errorf("not supported")
	case FilterLessOrEqual:
		return false, fmt.Errorf("not supported")
	case FilterPresent:
		attribute := strings.ToLower(f.Data.String())
		for k := range entry.Attributes {
			if strings.ToLower(k) == attribute {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("unsupported filter %v requested", f.Tag)
	}
	return false, nil
}
