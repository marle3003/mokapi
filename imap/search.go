package imap

import "strings"

type SearchRequest struct {
	Header  []SearchField
	NotFlag []string
}

type SearchField struct {
	Field string
	Value string
}

func (c *conn) Search(tag, d *Decoder) {
	r := &SearchRequest{}

	key := strings.ToUpper(d.Atom())
	switch key {
	case "UNSEEN":
		r.NotFlag = append(r.NotFlag, key[2:])
	}
}
