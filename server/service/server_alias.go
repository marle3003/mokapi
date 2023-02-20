package service

import (
	"fmt"
	"mokapi/sortedmap"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type ServerAliases struct {
	aliases *sortedmap.LinkedHashMap
}

func NewServerAliases() *ServerAliases {
	return &ServerAliases{aliases: sortedmap.NewLinkedHashMap()}
}

func (sa *ServerAliases) Parse(values []string) error {
	for _, alias := range values {
		kv := strings.Split(alias, "=")
		if len(kv) != 2 {
			return fmt.Errorf("expected server alias with format key=values, got: %v", alias)
		}
		values := strings.FieldsFunc(kv[1], func(r rune) bool {
			return r == ' ' || r == ','
		})
		aliases := make([]string, 0, len(values))
		for _, value := range values {
			aliases = append(aliases, value)
		}
		sa.aliases.Set(kv[0], aliases)
	}
	return nil
}

func (sa *ServerAliases) MatchAny(host string, r *http.Request) bool {
	requestHost, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		requestHost = r.Host
	}

	if v := sa.aliases.Get(host); v != nil {
		aliases := v.([]string)
		for _, alias := range aliases {
			if sa.matchAlias(requestHost, alias) {
				return true
			}
		}
	}
	if v := sa.aliases.Get("*"); v != nil {
		aliases := v.([]string)
		for _, alias := range aliases {
			if sa.matchAlias(requestHost, alias) {
				return true
			}
		}
	}
	return false
}

func (sa *ServerAliases) matchAlias(host, alias string) bool {
	u, err := url.Parse(alias)
	if len(u.Host) == 0 {
		u, _ = url.Parse("//" + alias)
	}
	_ = err
	if alias == "*" {
		return true
	}

	if strings.ToLower(alias) == strings.ToLower(host) {
		return true
	}

	return false
}
