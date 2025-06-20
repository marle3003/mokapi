package media

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mime"
	"reflect"
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
	c.Type, c.Parameters, _ = mime.ParseMediaType(s)
	if q, ok := c.Parameters["q"]; ok {
		var err error
		if c.Q, err = strconv.ParseFloat(q, 64); err != nil {
			log.Debugf("invalid q parameter in %v", s)
			c.Q = 1.0
		}
	}
	t := strings.Split(c.Type, "/")
	if len(t) > 1 {
		c.Type = strings.TrimSpace(t[0])
		c.Subtype = strings.TrimSpace(t[1])
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

func (c ContentType) IsAny() bool {
	return c.Type == "" || c.Type == "*" && c.Subtype == "*"
}

func (c ContentType) IsRange() bool {
	return c.Type != "*" && c.Subtype == "*"
}

func (c ContentType) IsEmpty() bool {
	return c.raw == ""
}

func (c ContentType) IsPrecise() bool {
	return !c.IsEmpty() && c.Type != "*" && c.Subtype != "*"
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
				return reflect.DeepEqual(c.Parameters, other.Parameters)
			}
			return true
		}
	}
	return false
}

func (c ContentType) IsXml() bool {
	return c.Subtype == "xml" || c.Subtype == "rss+xml" || c.Subtype == "xhtml+xml" || c.Subtype == "atom+xml" || c.Subtype == "xslt+xml"
}

func (c ContentType) IsDerivedFrom(contentType ContentType) bool {
	if contentType.IsAny() {
		return true
	}
	if !contentType.IsRange() {
		return c.Match(contentType)
	}
	return c.Type == contentType.Type
}

func Equal(c1, c2 ContentType) bool {
	return c1.raw == c2.raw
}
