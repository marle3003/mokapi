package directory

import (
	"bytes"
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"mokapi/ldap"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"strconv"
	"strings"
	"time"
)

type predicate func(entry Entry) bool

func (d *Directory) serveSearch(rw ldap.ResponseWriter, r *ldap.Request) {
	msg := r.Message.(*ldap.SearchRequest)
	m, doMonitor := monitor.LdapFromContext(r.Context)
	event := NewLogEvent(msg, events.NewTraits().WithName(d.config.Info.Name))
	defer func() {
		i := r.Context.Value("time")
		if i != nil {
			t := i.(time.Time)
			event.Duration = time.Now().Sub(t).Milliseconds()
		}
	}()

	log.Infof("ldap search request: messageId=%v BaseDN=%v Filter=%v",
		r.MessageId, msg.BaseDN, msg.Filter)

	if doMonitor {
		m.Search.WithLabel(d.config.Info.Name).Add(1)
		m.LastSearch.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	n := int64(0)
	sizeLimit := msg.SizeLimit
	pageLimit, pagedStoredIndex := getPageInfo(msg.Controls, r.Context)
	skipPageIndex := int64(0)
	maxSizeLimit := d.config.getSizeLimit()
	var results []ldap.SearchResult
	predicate, pos, err := compileFilter(msg.Filter)
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
	for _, e := range d.config.Entries {
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
		res.Attributes["objectClass"] = e.Attributes["objectClass"]

		if len(msg.Attributes) > 0 {
			for _, a := range msg.Attributes {
				for k, v := range e.Attributes {
					if strings.ToLower(a) == strings.ToLower(k) {
						res.Attributes[a] = v
					}
				}
			}
		} else {
			res.Attributes = e.Attributes
		}

		log.Debugf("found result for message %v: %v", r.MessageId, res.Dn)
		results = append(results, res)
		event.Response.Results = append(event.Response.Results, LdapSearchResult{
			Dn:         res.Dn,
			Attributes: res.Attributes,
		})
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

func compileFilter(filter string) (predicate, int, error) {
	if len(filter) == 0 || filter[0] != '(' {
		return nil, 0, fmt.Errorf("filter syntax error: expected starting with ( got %v", filter)
	}

	var attr *bytes.Buffer
	var v *bytes.Buffer
	var op int
	for pos := 0; pos < len(filter); pos++ {
		c := rune(filter[pos])
		switch {
		case c == '(':
			v = bytes.NewBuffer(nil)
		case c == ')':
			return newPredicate(op, attr.String(), v.String()), pos + 1, nil
		case c == '=':
			op = ldap.FilterEqualityMatch
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '>' && filter[pos+1] == '=':
			pos++
			op = ldap.FilterGreaterOrEqual
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '<' && filter[pos+1] == '=':
			pos++
			op = ldap.FilterLessOrEqual
			attr = v
			v = bytes.NewBuffer(nil)
		case c == '!':
			f, n, err := compileFilter(filter[pos+1:])
			return not(f), pos + n + 2, err
		case c == '&':
			fs, n, err := compileFilterSet(filter[pos+1:])
			return and(fs), pos + n + 2, err
		case c == '|':
			fs, n, err := compileFilterSet(filter[pos+1:])
			return or(fs), pos + n + 2, err
		default:
			v.WriteRune(c)
		}
	}

	return nil, 0, fmt.Errorf("unexpected filter end: %v", filter)
}

func compileFilterSet(filter string) ([]predicate, int, error) {
	pos := 0
	var fs []predicate
	for pos < len(filter) && filter[pos] != ')' {
		f, n, err := compileFilter(filter[pos:])

		if err != nil {
			return nil, 0, err
		}
		fs = append(fs, f)
		pos += n
	}
	return fs, pos, nil
}

func newPredicate(op int, name, value string) predicate {
	switch op {
	case ldap.FilterEqualityMatch:
		if strings.Contains(value, "*") {
			return substring(name, value)
		}
		return equal(name, value)
	case ldap.FilterGreaterOrEqual:
		return greaterOrEqual(name, value)
	case ldap.FilterLessOrEqual:
		return lessOrEqual(name, value)
	default:
		log.Errorf("unsupported filter operation for attribute %v: %v", name, op)
		return func(e Entry) bool {
			return false
		}
	}
}

func substring(name, value string) predicate {
	name = strings.ToLower(name)

	if value == "*" {
		return func(e Entry) bool {
			for k := range e.Attributes {
				// special case for objectClass
				// Matches all entries, even if objectClass is not explicitly present,
				// because objectClass is fundamental to every LDAP entry by definition.
				if name == "objectclass" {
					return true
				}
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

	return check(name, match)
}

func equal(name, value string) predicate {
	f := func(s string) bool {
		return value == s
	}
	return check(name, f)
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

func greaterOrEqual(name, value string) predicate {
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
	return check(name, f)
}

func lessOrEqual(name, value string) predicate {
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
	return check(name, f)
}

func check(name string, f func(string) bool) predicate {
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
