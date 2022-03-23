package media

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

var Empty = ContentType{}
var Any = ParseContentType("*/*")
var Default = ParseContentType("application/json")

type ContentType struct {
	Type       string
	Subtype    string
	Parameters map[string]string
	raw        string
	Q          float64
}

func ParseContentType(s string) ContentType {
	c := ContentType{raw: s, Parameters: make(map[string]string), Q: 1.0}
	a := strings.Split(s, ";")
	m := strings.Split(a[0], "/")
	c.Type = strings.ToLower(strings.TrimSpace(m[0]))
	if len(m) > 1 {
		c.Subtype = strings.ToLower(strings.TrimSpace(m[1]))
	}
	for _, p := range a[1:] {
		kv := strings.Split(p, "=")
		switch kv[0] {
		case "q":
			if len(kv) > 1 {
				var err error
				if c.Q, err = strconv.ParseFloat(kv[1], 64); err != nil {
					log.Debugf("invalid q parameter in %v", s)
					c.Q = 1.0
				}
			}
		default:
			if len(kv) > 1 {
				c.Parameters[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			} else {
				c.Parameters[kv[0]] = ""
			}
		}
	}

	return c
}

func (c ContentType) Key() string {
	if len(c.Subtype) > 0 {
		return fmt.Sprintf("%v/%v", c.Type, c.Subtype)
	}
	return c.Type
}

func (c ContentType) String() string {
	return c.raw
}

func (c ContentType) IsRange() bool {
	return c.Type == "" || c.Type == "*" || c.Subtype == "*"
}

func (c ContentType) IsEmpty() bool {
	return c.raw == ""
}

func (c ContentType) Match(other ContentType) bool {
	if c.Type == "*" || other.Type == "*" {
		return true
	}
	if c.Type == other.Type {
		if c.Subtype == "*" || other.Subtype == "*" {
			return true
		}
		if c.Subtype == other.Subtype {
			if len(c.Parameters) > 0 && len(other.Parameters) > 0 {
				return c.raw == other.raw
			}
			return true
		}
	}
	return false
}

func Equal(c1, c2 ContentType) bool {
	return c1.raw == c2.raw
}
