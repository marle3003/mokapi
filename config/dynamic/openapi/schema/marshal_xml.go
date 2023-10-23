package schema

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mokapi/sortedmap"
	"mokapi/xml"
	"net/url"
	"strings"
)

func marshalXml(i interface{}, r *Ref) ([]byte, error) {
	if r == nil || r.Value == nil {
		return nil, fmt.Errorf("no schema provided")
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	var name string
	if r.Value.Xml != nil && len(r.Value.Xml.Name) > 0 {
		name = r.Value.Xml.Name
	} else if len(r.Ref) > 0 {
		u, _ := url.Parse(r.Ref)
		seg := strings.Split(u.Fragment, "/")
		name = seg[len(seg)-1]
	} else {
		return nil, fmt.Errorf("root element name is undefined: reference name and xml.name is empty")
	}

	if i == nil {
		_, _ = writer.WriteString(fmt.Sprintf("<%v></%v>", name, name))
	} else {
		writeXmlElement(name, i, r, writer)
	}

	err := writer.Flush()
	return buffer.Bytes(), err
}

func writeXmlElement(name string, i interface{}, r *Ref, w io.Writer) {
	var s *Schema
	if r != nil {
		s = r.Value
	}

	node := xml.NewNode(name)
	wrapped := false
	if s != nil && s.Xml != nil {
		wrapped = s.Xml.Wrapped
		if len(s.Xml.Name) > 0 {
			node.Name = s.Xml.Name
		}
		if s.Xml.Prefix != "" {
			node.Name = fmt.Sprintf("%s:%s", s.Xml.Prefix, node.Name)
		}
		if s.Xml.Namespace != "" {
			attrName := "xmlns"
			if len(s.Xml.Prefix) > 0 {
				attrName += fmt.Sprintf(":%v", s.Xml.Prefix)
			}
			node.Attributes.Set(attrName, s.Xml.Namespace)
		}
	}

	switch v := i.(type) {
	case []interface{}:
		if wrapped {
			node.WriteStart(w)
		}
		for _, item := range v {
			if item == nil {
				continue
			}
			writeXmlElement(node.Name, item, s.Items, w)
		}
		if wrapped {
			node.WriteEnd(w)
		}
	case *marshalObject:
		setAttributes(node, v.LinkedHashMap, r)
		node.WriteStart(w)

		for it := v.LinkedHashMap.Iter(); it.Next(); {
			if it.Value() == nil {
				continue
			}

			propName := it.Key()
			prop := r.getProperty(propName)
			if prop != nil {
				x := prop.getXml()
				if x != nil && x.Attribute {
					continue
				}
			}

			writeXmlElement(propName, it.Value(), prop, w)
		}

		node.WriteEnd(w)
	default:
		node.Content = fmt.Sprintf("%v", i)
		node.Write(w)
	}
}

func setAttributes(n *xml.Node, m *sortedmap.LinkedHashMap[string, interface{}], r *Ref) {
	for it := m.Iter(); it.Next(); {
		if it.Value() == nil {
			continue
		}

		x := r.getPropertyXml(it.Key())
		if x == nil || !x.Attribute {
			continue
		}

		attrName := it.Key()
		if len(x.Name) > 0 {
			attrName = x.Name
		}
		if len(x.Prefix) > 0 {
			attrName = fmt.Sprintf("%s:%s", x.Prefix, attrName)
		}

		n.Attributes.Set(attrName, fmt.Sprintf("%v", it.Value()))
	}
}

func unmarshalXml(b []byte, r *Ref) (interface{}, error) {
	n, err := xml.Read(b)
	if err != nil {
		return nil, err
	}
	if len(n.Children) == 0 {
		return nil, nil
	}
	return parseXml(n, r)
}

func parseXml(n *xml.Node, r *Ref) (interface{}, error) {
	if len(n.Content) > 0 {
		return n.Content, nil
	}

	if r == nil || r.Value == nil {
		return parseFreeForm(n)
	}

	s := r.Value
	switch s.Type {
	case "object":
		if s.Properties == nil && s.IsFreeForm() {
			return parseFreeForm(n)
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
				v, err := parseXml(c, prop)
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
			v, err := parseXml(item, s.Items)
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

func parseFreeForm(n *xml.Node) (interface{}, error) {
	if isArray(n) {
		result := make([]interface{}, 0, len(n.Children))
		for _, c := range n.Children {
			v, err := parseXml(c, nil)
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
			v, err := parseXml(c, nil)
			if err != nil {
				return nil, err
			}
			result.Set(c.Name, v)
		}
		return result, nil
	}
}

func isArray(n *xml.Node) bool {
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
