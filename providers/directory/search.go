package directory

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"math"
	"mokapi/ldap"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type predicate func(entry Entry) bool

func (d *Directory) serveSearch(rw ldap.ResponseWriter, r *ldap.Request) {
	msg := r.Message.(*ldap.SearchRequest)
	m, doMonitor := monitor.LdapFromContext(r.Context)
	event := NewSearchLogEvent(msg, events.NewTraits().WithName(d.config.Info.Name))
	defer func() {
		i := r.Context.Value("time")
		if i != nil {
			t := i.(time.Time)
			event.Duration = time.Now().Sub(t).Milliseconds()
		}
	}()

	log.Infof("ldap search request: messageId=%v, Scope=%v BaseDN=%v Filter=%v",
		r.MessageId, msg.Scope, msg.BaseDN, msg.Filter)

	if doMonitor {
		m.RequestCounter.WithLabel(d.config.Info.Name, "search").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	n := int64(0)
	sizeLimit := msg.SizeLimit
	pageLimit, pagedStoredIndex := getPageInfo(msg.Controls, r.Context)
	skipPageIndex := int64(0)
	maxSizeLimit := d.config.getSizeLimit()
	var results []ldap.SearchResult
	p := &parser{s: d.config.Schema}
	predicate, pos, err := p.parse(msg.Filter)
	if pos != len(msg.Filter) || err != nil {
		if err != nil {
			log.Errorf("ldap: filter syntax error: %v", err)
		} else {
			log.Errorf("ldap: unsupported filter: %v", msg.Filter)
		}
		_ = rw.Write(&ldap.SearchResponse{Status: ldap.OperationsError})
		return
	}
	status := ldap.Success
	if d.config.Entries != nil {
		for it := d.config.Entries.Iter(); it.Next(); {
			e := it.Value()
			if !predicate(e) {
				continue
			}

			switch msg.Scope {
			case ldap.ScopeBaseObject:
				if e.Dn != msg.BaseDN {
					continue
				}
			case ldap.ScopeSingleLevel:
				parts := strings.Split(e.Dn, ",")
				if len(parts) < 2 && e.Dn != msg.BaseDN {
					continue
				}
				if dn := strings.Join(parts[1:], ","); dn != msg.BaseDN {
					continue
				}
			}
			if d.skip(&e, msg.BaseDN) {
				continue
			}

			if pagedStoredIndex != 0 && skipPageIndex < pagedStoredIndex {
				skipPageIndex++
				continue
			}
			if sizeLimit != 0 && n >= sizeLimit {
				break
			}
			if pageLimit > 0 && n >= pageLimit {
				setPageCookie(msg.Controls, n, r.Context)
				break
			}
			if maxSizeLimit > 0 && n >= maxSizeLimit {
				log.Errorf("ldap search query %v: size limit exceeded", msg.Filter)
				status = ldap.SizeLimitExceeded
				break
			}
			n++

			res := ldap.NewSearchResult(e.Dn)
			res.Attributes = getAttributes(msg.Attributes, &e)

			log.Debugf("found result for message %v: %v", r.MessageId, res.Dn)
			results = append(results, res)
			event.Response.Results = append(event.Response.Results, SearchResult{
				Dn:         res.Dn,
				Attributes: res.Attributes,
			})
		}
	}

	res := &ldap.SearchResponse{
		Status:   status,
		Results:  results,
		Message:  ldap.StatusText[status],
		Controls: msg.Controls,
	}

	event.Response.Status = ldap.StatusText[status]
	event.Actions = d.emitter.Emit("ldap", msg, res)

	if err := rw.Write(res); err != nil {
		log.Errorf("ldap: send search done: %v", err)
	}
}

func getAttributes(attr []string, e *Entry) map[string][]string {
	result := make(map[string][]string)
	plus := slices.Contains(attr, "+")
	star := slices.Contains(attr, "*") || len(attr) == 0

	if slices.Contains(attr, "1.1") {
		result["dn"] = []string{e.Dn}
		return result
	}

	for k, v := range e.Attributes {
		switch k {
		case "subschemaSubentry",
			"namingContexts",
			"objectClasses":
			if plus || slices.Contains(attr, k) {
				result[k] = v
			}
		default:
			if star || slices.Contains(attr, k) {
				result[k] = v
			}
		}
	}

	return result
}

type parser struct {
	s *Schema
}

func (p *parser) parse(filter string) (predicate, int, error) {
	if len(filter) == 0 || filter[0] != '(' {
		return nil, 0, fmt.Errorf("filter syntax error: expected starting with ( got %v", filter)
	}

	var attr *bytes.Buffer
	var v *bytes.Buffer
	var op int
	for pos := 0; pos < len(filter); pos++ {
		c := filter[pos]
		switch {
		case c == '(':
			v = bytes.NewBuffer(nil)
		case c == ')':
			p, err := p.predicate(op, attr.String(), v.String())
			return p, pos + 1, err
		case c == '=' && op == 0:
			op = ldap.FilterEqualityMatch
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '>' && filter[pos+1] == '=' && op == 0:
			pos++
			op = ldap.FilterGreaterOrEqual
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '<' && filter[pos+1] == '=' && op == 0:
			pos++
			op = ldap.FilterLessOrEqual
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '~' && filter[pos+1] == '=' && op == 0:
			pos++
			op = ldap.FilterApproxMatch
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '!':
			f, n, err := p.parse(filter[pos+1:])
			return not(f), pos + n + 2, err
		case c == '&':
			fs, n, err := p.parseSet(filter[pos+1:])
			return and(fs), pos + n + 2, err
		case c == '|':
			fs, n, err := p.parseSet(filter[pos+1:])
			return or(fs), pos + n + 2, err

		default:
			v.WriteByte(c)
		}
	}

	return nil, 0, fmt.Errorf("unexpected filter end: %v", filter)
}

func (p *parser) parseSet(filter string) ([]predicate, int, error) {
	pos := 0
	var fs []predicate
	for pos < len(filter) && filter[pos] != ')' {
		f, n, err := p.parse(filter[pos:])

		if err != nil {
			return nil, 0, err
		}
		fs = append(fs, f)
		pos += n
	}
	return fs, pos, nil
}

func (p *parser) predicate(op int, name, value string) (predicate, error) {
	switch op {
	case ldap.FilterEqualityMatch:
		if strings.Contains(value, "*") {
			return p.substring(name, value), nil
		}
		return p.equal(name, value)
	case ldap.FilterGreaterOrEqual:
		return p.greaterOrEqual(name, value), nil
	case ldap.FilterLessOrEqual:
		return p.lessOrEqual(name, value), nil
	case ldap.FilterApproxMatch:
		return p.check(name, func(s string) bool {
			n := len(s)

			if n > 5 && strings.Contains(s, value) {
				return true
			}
			distance := levenshtein(value, s)
			threshold := n / 5
			return distance <= threshold
		}), nil
	default:
		return nil, fmt.Errorf("unsupported filter operation for attribute %v: %v", name, op)
	}
}

func (p *parser) substring(name, value string) predicate {
	name = strings.ToLower(name)

	if value == "*" {
		return func(e Entry) bool {
			// special case for objectClass
			// Matches all entries, even if objectClass is not explicitly present,
			// because objectClass is fundamental to every LDAP entry by definition.
			if name == "objectclass" {
				return true
			}

			for k := range e.Attributes {
				if strings.ToLower(k) == name {
					return true
				}
			}
			return false
		}
	}

	parts := strings.Split(value, "*")
	var fs []func(string) bool
	for i, part := range parts {
		part := part
		if len(part) == 0 {
			continue
		}
		var f func(string, string) bool
		switch i {
		case 0:
			f = strings.HasPrefix
		case len(parts) - 1:
			f = strings.HasSuffix
		default:
			f = strings.Contains
		}
		fs = append(fs, func(s string) bool {
			return f(s, part)
		})
	}
	match := func(s string) bool {
		for _, f := range fs {
			if !f(s) {
				return false
			}
		}
		return true
	}

	return p.check(name, match)
}

func (p *parser) equal(name, value string) (predicate, error) {
	f := func(s string) bool {
		return value == s
	}

	if p.s != nil {
		t, ok := p.s.AttributeTypes[name]
		if ok {
			switch t.Equality {
			case "caseIgnoreMatch", "2.5.13.2":
				f = func(s string) bool {
					return strings.EqualFold(value, s)
				}
			case "caseExactMatch", "2.5.13.5":
				f = func(s string) bool {
					return value == s
				}
			case "integerMatch", "2.5.13.14":
				v, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid integer value %v: %v", value, err)
				}
				f = func(s string) bool {
					i, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						return false
					}
					return i == v
				}
			case "octetStringMatch", "2.5.13.17":
				v, err := base64.StdEncoding.DecodeString(value)
				if err != nil {
					return nil, fmt.Errorf("invalid octet value %v: %v", value, err)
				}
				f = func(s string) bool {
					b := []byte(s)
					return bytes.Equal(b, v)
				}
			case "booleanMatch", "2.5.13.13":
				f = func(s string) bool {
					return value == s
				}
			case "numericStringMatch", "2.5.13.8":
				if !isNumericString(value) {
					f = func(s string) bool { return false }
				} else {
					f = func(s string) bool {
						if !isNumericString(s) {
							return false
						}
						return value == s
					}
				}
			case "distinguishedNameMatch", "2.5.13.1":
				f = func(s string) bool {
					return strings.EqualFold(value, s)
				}
			case "telephoneNumberMatch", "2.5.13.20":
				v := strings.ReplaceAll(value, " ", "")
				v = strings.ReplaceAll(v, "-", "")
				f = func(s string) bool {
					s = strings.ReplaceAll(value, " ", "")
					s = strings.ReplaceAll(v, "-", "")
					return v == s
				}
			default:
				return nil, fmt.Errorf("unsupported equality type: %v", t)
			}
		}
	}
	return p.check(name, f), nil
}

func and(fs []predicate) predicate {
	return func(e Entry) bool {
		for _, f := range fs {
			if !f(e) {
				return false
			}
		}
		return true
	}
}

func or(fs []predicate) predicate {
	return func(e Entry) bool {
		for _, f := range fs {
			if f(e) {
				return true
			}
		}
		return false
	}
}

func not(f predicate) predicate {
	return func(e Entry) bool {
		return !f(e)
	}
}

func (p *parser) greaterOrEqual(name, value string) predicate {
	n, err := strconv.ParseFloat(value, 64)
	var f func(string) bool
	if err != nil {
		f = func(s string) bool {
			return s >= value
		}
	} else {
		f = func(s string) bool {
			toCompare, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return false
			}
			return toCompare >= n
		}
	}
	return p.check(name, f)
}

func (p *parser) lessOrEqual(name, value string) predicate {
	n, err := strconv.ParseFloat(value, 64)
	var f func(string) bool
	if err != nil {
		f = func(s string) bool {
			return s <= value
		}
	} else {
		f = func(s string) bool {
			toCompare, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return false
			}
			return toCompare <= n
		}
	}
	return p.check(name, f)
}

func (p *parser) check(name string, f func(string) bool) predicate {
	return func(e Entry) bool {
		for k, attrs := range e.Attributes {
			if strings.ToLower(name) != strings.ToLower(k) {
				continue
			}
			for _, attr := range attrs {
				if f(attr) {
					return true
				}
			}
		}
		return false
	}
}

func getPageInfo(controls []ldap.Control, ctx context.Context) (int64, int64) {
	for _, control := range controls {
		switch ctrl := control.(type) {
		case *ldap.PagedResultsControl:
			page := ldap.PagingFromContext(ctx)
			pagedIndex, ok := page.Cookies[ctrl.Cookie]
			if !ok {
				pagedIndex = 0
			}
			return ctrl.PageSize, pagedIndex
		}
	}
	return -1, 0
}

func setPageCookie(controls []ldap.Control, pageIndex int64, ctx context.Context) {
	for _, control := range controls {
		switch ctrl := control.(type) {
		case *ldap.PagedResultsControl:
			name := gofakeit.LetterN(6)
			ctrl.Cookie = name
			page := ldap.PagingFromContext(ctx)
			page.Cookies[name] = pageIndex
		}
	}
}

func isNumericString(s string) bool {
	for _, ch := range s {
		if !unicode.IsNumber(ch) {
			return false
		}
	}
	return true
}

// Levenshtein function calculates the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	// Create a matrix to hold the distances
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize the first row and column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			substitutionCost := 0
			if a[i-1] != b[j-1] {
				substitutionCost = 1
			}
			// get the smallest of three numbers
			matrix[i][j] = int(math.Min(
				float64(matrix[i-1][j]+1), // Deletion
				math.Min(
					float64(matrix[i][j-1]+1), // Insertion
					float64(matrix[i-1][j-1]+substitutionCost),
				), // Substitution
			))
		}
	}

	// Return the final Levenshtein distance
	return matrix[len(a)][len(b)]
}
