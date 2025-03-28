package schema

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []node     `xml:",any"`
}

func (n *node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type proxy node

	return d.DecodeElement((*proxy)(n), &start)
}

func UnmarshalXML(r io.Reader, ref *Schema) (interface{}, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}

	n := &node{}
	err = xml.Unmarshal(b, &n)
	if err != nil {
		return nil, err
	}
	return parse(n, ref)
}

func parse(n *node, s *Schema) (interface{}, error) {
	if len(n.Nodes) == 0 && len(n.Attrs) == 0 {
		return parseValue(string(n.Content), s)
	}

	if isArray(n) || (s != nil && s.Type.IsArray()) {
		var items *Schema
		if s != nil {
			items = s.Items
		}
		if _, wrapped := isWrapped(s); wrapped {
			if len(n.Nodes) == 0 {
				return nil, nil
			}
			n = &n.Nodes[0]
		}
		return getItems(n, items)
	}

	m := map[string]interface{}{}
	for _, child := range n.Nodes {
		name, prop := getProperty(child.XMLName, s, false)
		v, err := parse(&child, prop)
		if err != nil {
			return nil, err
		}
		if _, ok := m[name]; ok {
			if arr, isArray := m[name].([]interface{}); isArray {
				m[name] = append(arr, v)
			} else {
				m[name] = []interface{}{m[name], v}
			}
		} else {
			m[name] = v
		}
	}
	for _, attr := range n.Attrs {
		name, prop := getProperty(attr.Name, s, true)
		v, err := parseValue(attr.Value, prop)
		if err != nil {
			return nil, err
		}
		m[name] = v
	}
	return m, nil
}

func parseValue(s string, ref *Schema) (interface{}, error) {
	if ref == nil || ref.Type.IsString() {
		return s, nil
	}

	t := ref.Type

	if t.IsInteger() {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("parse integer failed: %v", s)
		}
		return v, nil
	}

	if t.IsNumber() {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("parse floating number failed: %v", s)
		}
		return v, nil
	}

	if t.IsBool() {
		return s == "true", nil
	}

	return nil, fmt.Errorf("unknown type: %v", ref.Type)
}

func getProperty(name xml.Name, s *Schema, asAttr bool) (string, *Schema) {
	if s == nil || !s.HasProperties() {
		return name.Local, nil
	}

	prop := s.Properties.Get(name.Local)
	if prop != nil {
		if prop.Xml != nil {
			x := prop.Xml
			if len(x.Prefix) > 0 && x.Prefix == name.Space {
				return name.Local, prop
			}
		} else {
			return name.Local, prop
		}

	}

	for it := s.Properties.Iter(); it.Next(); {
		prop = it.Value()
		if prop == nil {
			continue
		}
		x := prop.Xml
		if x == nil {
			continue
		}
		if x.Name == name.Local && x.Attribute == asAttr {
			return it.Key(), prop
		}
	}

	return name.Local, nil
}

func getItems(n *node, ref *Schema) ([]interface{}, error) {
	var r []interface{}
	for _, child := range n.Nodes {
		v, err := parse(&child, ref)
		if err != nil {
			return nil, err
		}
		r = append(r, v)
	}
	return r, nil
}

func isArray(n *node) bool {
	if len(n.Nodes) <= 1 {
		return false
	}
	name := n.Nodes[0].XMLName.Local
	for _, child := range n.Nodes[1:] {
		if child.XMLName.Local != name {
			return false
		}
	}
	return true
}

func isWrapped(ref *Schema) (string, bool) {
	if ref == nil || ref.Xml == nil {
		return "", false
	}
	x := ref.Xml
	return x.Name, x.Wrapped
}
