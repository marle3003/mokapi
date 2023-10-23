package xml

import (
	"bytes"
	"encoding/xml"
	"strings"
)

func Read(b []byte) (*Node, error) {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	n := NewNode("")
	err := n.decode(decoder, nil)
	return n, err
}

func (n *Node) decode(d *xml.Decoder, start *xml.StartElement) error {
	if start == nil {
		for {
			tok, err := d.Token()
			if err != nil {
				return err
			}
			if t, ok := tok.(xml.StartElement); ok {
				start = &t
				break
			}
		}
	}

	n.Name = start.Name.Local

	for _, a := range start.Attr {
		n.Attributes.Set(a.Name.Local, a.Value)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			child := NewNode("")
			err = child.decode(d, &t)
			if err != nil {
				return err
			}

			n.Children = append(n.Children, child)
		case xml.EndElement:
			return nil
		case xml.CharData:
			s := string(t)
			n.Content = strings.TrimSuffix(strings.TrimSpace(s), "\n")
		}
	}
}
