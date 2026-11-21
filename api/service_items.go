package api

import (
	"mokapi/providers/asyncapi3"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"slices"
	"strings"
)

// serviceItem is a compact entry of a service (endpoint, topic, channel,
// mailbox) used by the dashboard to filter services without further requests.
type serviceItem struct {
	Name   string   `json:"name"`
	Method string   `json:"method,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

var httpMethodOrder = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "QUERY"}

func getChannelItems(cfg *asyncapi3.Config, protocol string) []serviceItem {
	var items []serviceItem
	for _, ch := range cfg.Channels {
		if ch.Value == nil || !ch.Value.IsChannelAvailable(protocol) {
			continue
		}
		item := serviceItem{Name: ch.Value.ResolveAddress()}
		for _, tag := range ch.Value.Tags {
			if tag.Value != nil {
				item.Tags = append(item.Tags, tag.Value.Name)
			}
		}
		items = append(items, item)
	}
	sortServiceItems(items)
	return items
}

func sortServiceItems(items []serviceItem) {
	slices.SortFunc(items, func(a, b serviceItem) int {
		if c := strings.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return compareHttpMethods(a.Method, b.Method)
	})
}

// compareHttpMethods orders the usual methods as in httpMethodOrder and other
// methods after them, alphabetically.
func compareHttpMethods(a, b string) int {
	if c := methodRank(a) - methodRank(b); c != 0 {
		return c
	}
	return strings.Compare(a, b)
}

func methodRank(method string) int {
	if i := slices.Index(httpMethodOrder, strings.ToUpper(method)); i >= 0 {
		return i
	}
	return len(httpMethodOrder)
}

func getHttpItems(hi *runtime.HttpInfo) []serviceItem {
	var items []serviceItem
	if hi.Paths == nil {
		return items
	}
	for it := hi.Paths.Iter(); it.Next(); {
		p := it.Value()
		if p.Value == nil {
			continue
		}
		for method, o := range p.Value.Operations() {
			items = append(items, serviceItem{
				Name:   it.Key(),
				Method: strings.ToLower(method),
				Tags:   o.Tags,
			})
		}
	}
	sortServiceItems(items)
	return items
}

func getMailboxItems(s *mail.Store) []serviceItem {
	mailboxes := s.Mailboxes()
	items := make([]serviceItem, 0, len(mailboxes))
	for _, mb := range mailboxes {
		items = append(items, serviceItem{Name: mb.Name})
	}
	sortServiceItems(items)
	return items
}
