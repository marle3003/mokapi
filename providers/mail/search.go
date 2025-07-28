package mail

import (
	"context"
	"fmt"
	"mokapi/imap"
	"strings"
)

func (h *Handler) Search(req *imap.SearchRequest, ctx context.Context) (*imap.SearchResponse, error) {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	f := mb.Select(selected)
	if f == nil {
		return nil, fmt.Errorf("mailbox not found")
	}

	set := &imap.IdSet{}

	for i, m := range f.Messages {
		msn := uint32(i + 1)
		if match(msn, m, req.Criteria) {
			set.AddId(msn)
		}
	}

	return &imap.SearchResponse{All: set}, nil
}

func match(msn uint32, m *Mail, criteria *imap.SearchCriteria) bool {
	if criteria == nil {
		return true
	}

	if criteria.Seq != nil && !criteria.Seq.Contains(msn) {
		return false
	}

	for _, flag := range criteria.Flag {
		if !m.HasFlag(flag) {
			return false
		}
	}

	for _, field := range criteria.NotFlag {
		if m.HasFlag(field) {
			return false
		}
	}

	for _, header := range criteria.Headers {
		v := m.Message.Headers[header.Name]
		v = strings.ToLower(v)
		if !strings.Contains(v, header.Value) {
			return false
		}
	}

	if !criteria.Before.IsZero() && !m.Received.Before(criteria.Before) {
		return false
	}
	if !criteria.Since.IsZero() && !m.Received.After(criteria.Since) {
		return false
	}
	if !criteria.SentBefore.IsZero() && !m.Message.Date.Before(criteria.SentBefore) {
		return false
	}
	if !criteria.SentSince.IsZero() && !m.Message.Date.Before(criteria.SentSince) {
		return false
	}

	for _, body := range criteria.Body {
		if !strings.Contains(strings.ToLower(m.Body), body) {
			return false
		}
	}

	for _, text := range criteria.Text {
		found := false
		for _, val := range m.Headers {
			val = strings.ToLower(val)
			if strings.Contains(strings.ToLower(m.Body), text) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if criteria.Larger > 0 && int64(m.Size) <= criteria.Larger {
		return false
	}
	if criteria.Smaller > 0 && int64(m.Size) >= criteria.Larger {
		return false
	}

	for _, not := range criteria.Not {
		if match(msn, m, &not) {
			return false
		}
	}

	for _, or := range criteria.Or {
		if !match(msn, m, &or[0]) && !match(msn, m, &or[1]) {
			return false
		}
	}

	return true
}
