package ldap

import (
	"fmt"
	"mokapi/config/dynamic"
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

func (s *Server) handleSearchRequest(conn net.Conn, messageId int64, req *ber.Packet) error {
	searchRequest, error := parseSearchRequest(req)
	if error != nil {
		return error
	}

	log.Infof("Received search request with messageId %v BaseDN: %v Filter %v", messageId, searchRequest.baseObject, searchRequest.filterString)

	if searchRequest.baseObject == "" && searchRequest.scope == ScopeBaseObject {
		result := &SearchResult{dn: "", attributes: make(map[string][]string)}
		for k, v := range s.root.Attributes {
			result.attributes[k] = v
		}

		response := encodeSearchResult(messageId, result)
		sendResponse(conn, response)
		return nil
	}

	count := 0
	for _, entry := range s.entries {
		ok, error := filter(searchRequest.filter, entry)
		if error != nil {
			fmt.Print(error.Error())
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

		log.Infof("Found result for message %v: %v", messageId, result.dn)

		response := encodeSearchResult(messageId, result)
		sendResponse(conn, response)
	}

	log.Debugf("Finishing search for message %v", messageId)

	return nil
}

func filter(f *ber.Packet, entry *dynamic.Entry) (bool, error) {
	switch f.Tag {
	case FilterAnd:
		for _, child := range f.Children {
			ok, error := filter(child, entry)
			if error != nil {
				return false, error
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	case FilterOr:
		for _, child := range f.Children {
			ok, error := filter(child, entry)
			if error != nil {
				return false, error
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	case FilterNot:
		if len(f.Children) != 1 {
			return false, fmt.Errorf("Invalid filter operation")
		}

		ok, error := filter(f.Children[0], entry)
		if error != nil {
			return false, error
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
		return false, fmt.Errorf("Not supported")
	case FilterGreaterOrEqual:
		return false, fmt.Errorf("Not supported")
	case FilterLessOrEqual:
		return false, fmt.Errorf("Not supported")
	case FilterPresent:
		attribute := strings.ToLower(f.Data.String())
		for k := range entry.Attributes {
			if strings.ToLower(k) == attribute {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("Unsupported filter %v requested", f.Tag)
	}
	return false, nil
}
