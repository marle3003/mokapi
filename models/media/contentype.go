package media

import (
	"fmt"
	"strings"
)

type ContentType struct {
	Type       string
	Subtype    string
	Parameters map[string]string
	raw        string
}

func ParseContentType(s string) *ContentType {
	c := &ContentType{raw: s, Parameters: make(map[string]string)}
	a := strings.Split(s, ";")
	m := strings.Split(a[0], "/")
	c.Type = strings.ToLower(strings.TrimSpace(m[0]))
	if len(m) > 1 {
		c.Subtype = strings.ToLower(strings.TrimSpace(m[1]))
	}
	for _, p := range a[1:] {
		kv := strings.Split(p, "=")
		if len(kv) > 1 {
			c.Parameters[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			c.Parameters[kv[0]] = ""
		}
	}

	return c
}

func (c *ContentType) Key() string {
	if len(c.Subtype) > 0 {
		return fmt.Sprintf("%v/%v", c.Type, c.Subtype)
	}
	return c.Type
}

func (c *ContentType) String() string {
	return c.raw
}

func (c *ContentType) Equals(other *ContentType) bool {
	return c.Type == other.Type && c.Subtype == other.Subtype
}
