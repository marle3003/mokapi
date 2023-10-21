package schema

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"mokapi/sortedmap"
	"strings"
)

type Xml struct {
	Wrapped   bool   `yaml:"wrapped" json:"wrapped"`
	Name      string `yaml:"name" json:"name"`
	Attribute bool   `yaml:"attribute" json:"attribute"`
	Prefix    string `yaml:"prefix" json:"prefix"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

func parseXml(b []byte, s *Ref) (interface{}, error) {
	e, err := decode(b)
	if err != nil {
		return nil, err
	}
	if len(e.Children) == 0 {
		return nil, nil
	}
	return e.parse(s)
}

func decode(data []byte) (*xmlNode, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	e := &xmlNode{Children: make([]*xmlNode, 0), Attributes: &sortedmap.LinkedHashMap[string, string]{}}
	err := e.decode(decoder, nil)
	return e, err
}

func (n *xmlNode) parse(r *Ref) (interface{}, error) {
	if len(n.Content) > 0 {
		return n.Content, nil
	}

	if r == nil || r.Value == nil {
		return n.parseFreeForm()
	}

	s := r.Value
	switch s.Type {
	case "object":
		if s.Properties == nil && s.IsFreeForm() {
			return n.parseFreeForm()
		}
		props := sortedmap.NewLinkedHashMap()
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			xmlName := name
			prop := it.Value()
			if prop.Value.Xml != nil && len(prop.Value.Xml.Name) > 0 {
				xmlName = prop.Value.Xml.Name
			}
			if prop.Value.Xml != nil && prop.Value.Xml.Attribute {
				if v, found := n.Attributes.Get(xmlName); found {
					props.Set(name, v)
				}
			} else {
				c := n.GetFirstElement(xmlName)
				if c == nil {
					continue
				}
				v, err := c.parse(prop)
				if err != nil {
					return nil, err
				}
				props.Set(name, v)
			}
		}
		return props, nil
	case "array":
		if n == nil {
			return nil, fmt.Errorf("expected array but found null")
		}
		elements := n.Children
		if s.Xml != nil && s.Xml.Wrapped {
			if len(n.Children) == 0 {
				return make([]interface{}, 0), nil
			}
			elements = n.Children[0].Children
		}
		array := make([]interface{}, len(elements))
		for i, item := range n.Children {
			v, err := item.parse(s.Items)
			if err != nil {
				return nil, err
			}
			array[i] = v
		}
		return array, nil
	default:
		return n.Content, nil
	}
}

func (n *xmlNode) decode(d *xml.Decoder, start *xml.StartElement) error {
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
			child := &xmlNode{Children: make([]*xmlNode, 0), Attributes: &sortedmap.LinkedHashMap[string, string]{}}
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

func (n *xmlNode) parseFreeForm() (interface{}, error) {
	if n.IsArray() {
		result := make([]interface{}, 0, len(n.Children))
		for _, c := range n.Children {
			v, err := c.parse(nil)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
		return result, nil
	} else {
		result := sortedmap.NewLinkedHashMap()
		for it := n.Attributes.Iter(); it.Next(); {
			result.Set(it.Key(), it.Value())
		}
		for _, c := range n.Children {
			v, err := c.parse(nil)
			if err != nil {
				return nil, err
			}
			result.Set(c.Name, v)
		}
		return result, nil
	}
}

func (n *xmlNode) IsArray() bool {
	if len(n.Children) <= 1 {
		return false
	}
	name := n.Children[0].Name
	for i := 1; i < len(n.Children); i++ {
		if name != n.Children[i].Name {
			return false
		}
	}
	return true
}

func (n *xmlNode) GetFirstElement(name string) *xmlNode {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}
