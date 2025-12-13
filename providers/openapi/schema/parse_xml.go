package schema

import (
	"encoding/xml"
	"fmt"
	"io"
	"mokapi/schema/json/parser"
	"strconv"
)

type XmlParser struct {
	s *Schema
}

type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []node     `xml:",any"`
}

func NewXmlParser(s *Schema) *XmlParser {
	return &XmlParser{s: s}
}

func (p *XmlParser) Parse(v any) (any, error) {
	var b []byte
	switch vv := v.(type) {
	case string:
		b = []byte(vv)
	case []byte:
		b = vv
	default:
		return nil, fmt.Errorf("failed to parse XML: unsupported type: %T", v)
	}

	n := &node{}
	err := xml.Unmarshal(b, &n)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}
	data, err := parseXML(n, p.s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	pn := parser.Parser{
		Schema:                       ConvertToJsonSchema(p.s),
		ConvertStringToNumber:        true,
		ConvertStringToBoolean:       true,
		ValidateAdditionalProperties: true,
	}
	return pn.Parse(data)
}

func (p *XmlParser) ParseFrom(r io.Reader) (any, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return p.Parse(b)
}

func UnmarshalXML(r io.Reader, s *Schema) (interface{}, error) {
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
		return nil, fmt.Errorf("failed to unmarshal xml: %w", err)
	}
	return parseXML(n, s)
}

func parseXML(n *node, s *Schema) (any, error) {
	if len(n.Nodes) == 0 && len(n.Attrs) == 0 {
		if len(n.Content) == 0 {
			if s.Type.IsObject() {
				return map[string]any{}, nil
			}
			if s.Type.IsArray() {
				return []any{}, nil
			}
		}
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

	m := map[string]any{}

	// elements can override attribute values
	for _, attr := range n.Attrs {
		name, prop := getProperty(attr.Name, s)
		if prop != nil && prop.Xml != nil && !prop.Xml.Attribute {
			continue
		}
		v, err := parseValue(attr.Value, prop)
		if err != nil {
			return nil, err
		}
		m[name] = v
	}

	for _, child := range n.Nodes {
		name, prop := getProperty(child.XMLName, s)
		if prop != nil && prop.Xml != nil && prop.Xml.Attribute {
			continue
		}
		v, err := parseXML(&child, prop)
		if err != nil {
			return nil, err
		}
		if _, ok := m[name]; ok {
			if arr, isArray := m[name].([]any); isArray {
				m[name] = append(arr, v)
			} else {
				m[name] = []interface{}{m[name], v}
			}
		} else {
			m[name] = v
		}
	}

	if s != nil {
		for _, req := range s.Required {
			prop := s.Properties.Get(req)
			if prop == nil || prop.Xml == nil {
				// here we want to verify only XML
				continue
			}
			if prop.Xml.Attribute && !n.hasAttribute(req) {
				return nil, fmt.Errorf("required attribute '%s' not found in XML", req)
			}
			if !prop.Xml.Attribute && n.hasElement(req) {
				return nil, fmt.Errorf("required element '%s' not found in XML", req)
			}
		}
	}

	return m, nil
}

func parseValue(s string, ref *Schema) (interface{}, error) {
	if ref == nil || ref.Type.IsString() {
		return s, nil
	}

	t := ref.Type

	if t.IsInteger() {
		val64, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse integer failed: %v", s)
		}
		return int32(val64), nil
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

func getProperty(name xml.Name, s *Schema) (string, *Schema) {
	if s == nil || !s.HasProperties() {
		return name.Local, nil
	}

	for it := s.Properties.Iter(); it.Next(); {
		prop := it.Value()
		if prop == nil {
			continue
		}
		x := prop.Xml
		if x == nil {
			continue
		}
		if x.Name == name.Local {
			return it.Key(), prop
		}
	}

	prop := s.Properties.Get(name.Local)
	if prop != nil {
		if prop.Xml != nil {
			x := prop.Xml
			if x.Prefix == name.Space {
				return name.Local, prop
			}
		} else {
			return name.Local, prop
		}

	}

	return name.Local, nil
}

func getItems(n *node, ref *Schema) ([]interface{}, error) {
	var r []interface{}
	for _, child := range n.Nodes {
		v, err := parseXML(&child, ref)
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

func (n *node) hasAttribute(name string) bool {
	for _, attr := range n.Attrs {
		if attr.Name.Local == name {
			return true
		}
	}
	return false
}

func (n *node) hasElement(name string) bool {
	for _, elem := range n.Nodes {
		if elem.XMLName.Local == name {
			return true
		}
	}
	return false
}
