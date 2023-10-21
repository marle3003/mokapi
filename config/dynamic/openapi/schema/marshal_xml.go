package schema

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"mokapi/sortedmap"
	"net/url"
	"strings"
)

type xmlNode struct {
	Name       string
	Children   []*xmlNode
	Attributes *sortedmap.LinkedHashMap[string, string]
	Content    string
}

func (r *Ref) marshalXml(i interface{}) ([]byte, error) {
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

func writeXmlElement(name string, i interface{}, r *Ref, writer *bufio.Writer) {
	var s *Schema
	if r != nil {
		s = r.Value
	}

	node := &xmlNode{
		Name:       name,
		Attributes: &sortedmap.LinkedHashMap[string, string]{},
	}
	wrapped := false
	if s != nil && s.Xml != nil {
		wrapped = s.Xml.Wrapped && r.Value.Type == "array"
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

	if wrapped {
		_, _ = writer.WriteString(fmt.Sprintf("<%v>", node.Name))
	}

	switch v := i.(type) {
	case []interface{}:
		for _, item := range v {
			if item == nil {
				continue
			}
			writeXmlElement(node.Name, item, s.Items, writer)
		}
	case *marshalObject:
		node.setAttributes(v.LinkedHashMap, r)
		node.writeStart(writer)

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

			writeXmlElement(propName, it.Value(), prop, writer)
		}

		node.writeEnd(writer)
	default:
		node.writeStart(writer)
		text := fmt.Sprintf("%v", i)
		_ = xml.EscapeText(writer, []byte(text))
		node.writeEnd(writer)
	}

	if wrapped {
		_, _ = writer.WriteString(fmt.Sprintf("</%v>", node.Name))
	}
}

func (n *xmlNode) setAttributes(m *sortedmap.LinkedHashMap[string, interface{}], r *Ref) {
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

func (n *xmlNode) writeStart(writer *bufio.Writer) {
	_, _ = writer.WriteString("<")
	_, _ = writer.WriteString(n.Name)
	for it := n.Attributes.Iter(); it.Next(); {
		_, _ = writer.WriteString(fmt.Sprintf(" %v=\"%v\"", it.Key(), it.Value()))
	}
	_, _ = writer.WriteString(">")
}

func (n *xmlNode) writeEnd(writer *bufio.Writer) {
	_, _ = writer.WriteString(fmt.Sprintf("</%v>", n.Name))
}
