package schema

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

type xmlNode struct {
	Name       string
	Children   []*xmlNode
	Attributes map[string]string
	Content    string
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

func writeXml(i interface{}, r *Ref) ([]byte, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	name := ""
	if len(r.Ref) > 0 {
		u, _ := url.Parse(r.Ref)
		seg := strings.Split(u.Fragment, "/")
		name = seg[len(seg)-1]
	}

	err := encode(name, i, r, writer)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return buffer.Bytes(), nil
}

func encode(name string, i interface{}, r *Ref, writer *bufio.Writer) error {
	if r == nil || r.Value == nil {
		return fmt.Errorf("unable to encode xml without a schema")
	}

	s := r.Value
	node := &xmlNode{
		Name:       name,
		Attributes: make(map[string]string),
	}
	wrapped := false
	if s.Xml != nil {
		wrapped = s.Xml.Wrapped
		if len(s.Xml.Name) > 0 {
			node.Name = s.Xml.Name
		}
		if s.Xml.Prefix != "" {
			node.Name = fmt.Sprintf("%s:%s", s.Xml.Prefix, node.Name)
		}
		if s.Xml.Namespace != "" {
			node.Attributes["xmlns:"+s.Xml.Prefix] = s.Xml.Namespace
		}
	}

	if len(node.Name) == 0 {
		return fmt.Errorf("missing xml element name")
	}

	if wrapped {
		_, err := writer.WriteString(fmt.Sprintf("<%v>", node.Name))
		if err != nil {
			return err
		}
	}

	switch s.Type {
	case "array":
		if list, ok := i.([]interface{}); ok {
			for _, item := range list {
				err := encode(node.Name, item, s.Items, writer)
				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("expected array but got %T", i)
		}
	case "object":
		if i == nil {
			node.writeStart(writer)
			writer.WriteString(fmt.Sprintf("</%v>", node.Name))
		} else {
			o, ok := i.(*schemaObject)
			if !ok {
				return fmt.Errorf("expected object got %T", i)
			}

			for it := s.Properties.Value.Iter(); it.Next(); {
				prop := it.Value().(*Ref)
				if prop.Value == nil || prop.Value.Xml == nil || !prop.Value.Xml.Attribute {
					continue
				}
				x := prop.Value.Xml
				attrName := it.Key().(string)
				if len(x.Name) > 0 {
					attrName = x.Name
				}
				if len(x.Prefix) > 0 {
					attrName = fmt.Sprintf("%s:%s", x.Prefix, attrName)
				}
				v := o.Get(it.Key())
				node.Attributes[attrName] = fmt.Sprintf("%v", v)
			}

			node.writeStart(writer)

			for it := s.Properties.Value.Iter(); it.Next(); {
				prop := it.Value().(*Ref)
				if prop.Value == nil || (prop.Value.Xml != nil && prop.Value.Xml.Attribute) {
					continue
				}
				name := it.Key().(string)
				err := encode(name, o.Get(name), prop, writer)
				if err != nil {
					return err
				}
			}

			writer.WriteString(fmt.Sprintf("</%v>", node.Name))
		}
	default:
		node.writeStart(writer)

		if s.Xml != nil && s.Xml.CData {
			writer.WriteString(fmt.Sprintf("<![CDATA[%v]]>", i))
		} else {
			writer.WriteString(fmt.Sprintf("%v", i))
		}

		writer.WriteString(fmt.Sprintf("</%v>", node.Name))
	}

	if wrapped {
		writer.WriteString(fmt.Sprintf("</%v>", node.Name))
	}
	return nil
}

func decode(data []byte) (*xmlNode, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	e := &xmlNode{Children: make([]*xmlNode, 0), Attributes: make(map[string]string)}
	err := e.decode(decoder, nil)
	return e, err
}

func (e *xmlNode) parse(r *Ref) (interface{}, error) {
	if len(e.Content) > 0 {
		return e.Content, nil
	}

	if r == nil || r.Value == nil {
		return e.parseFreeForm()
	}

	s := r.Value
	switch s.Type {
	case "object":
		if s.IsFreeForm() {
			return e.parseFreeForm()
		}
		props := make(map[string]interface{})
		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			xmlName := name
			prop := it.Value().(*Ref)
			if prop.Value.Xml != nil && len(prop.Value.Xml.Name) > 0 {
				xmlName = prop.Value.Xml.Name
			}
			if prop.Value.Xml != nil && prop.Value.Xml.Attribute {
				if v, ok := e.Attributes[xmlName]; ok {
					props[name] = v
				}
			} else {
				c := e.GetFirstElement(xmlName)
				if c == nil {
					continue
				}
				v, err := c.parse(prop)
				if err != nil {
					return nil, err
				}
				props[name] = v
			}
		}
		return props, nil
	case "array":
		if e == nil {
			return nil, fmt.Errorf("expected array but found null")
		}
		elements := e.Children
		if s.Xml != nil && s.Xml.Wrapped {
			if len(e.Children) == 0 {
				return make([]interface{}, 0), nil
			}
			elements = e.Children[0].Children
		}
		array := make([]interface{}, len(elements))
		for i, item := range e.Children {
			v, err := item.parse(s.Items)
			if err != nil {
				return nil, err
			}
			array[i] = v
		}
		return array, nil
	default:
		return e.Content, nil
	}
}

func (e *xmlNode) decode(d *xml.Decoder, start *xml.StartElement) error {
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

	e.Name = start.Name.Local

	for _, a := range start.Attr {
		e.Attributes[a.Name.Local] = a.Value
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			child := &xmlNode{Children: make([]*xmlNode, 0), Attributes: make(map[string]string)}
			err = child.decode(d, &t)
			if err != nil {
				return err
			}

			e.Children = append(e.Children, child)
		case xml.EndElement:
			return nil
		case xml.CharData:
			s := string(t)
			e.Content = strings.TrimSuffix(strings.TrimSpace(s), "\n")
		}
	}
}

func (e *xmlNode) parseFreeForm() (interface{}, error) {
	if e.IsArray() {
		result := make([]interface{}, 0, len(e.Children))
		for _, c := range e.Children {
			v, err := c.parse(nil)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
		return result, nil
	} else {
		result := make(map[string]interface{})
		for n, a := range e.Attributes {
			result[n] = a
		}
		for _, c := range e.Children {
			v, err := c.parse(nil)
			if err != nil {
				return nil, err
			}
			result[c.Name] = v
		}
		return result, nil
	}
}

func (e *xmlNode) IsArray() bool {
	if len(e.Children) <= 1 {
		return false
	}
	name := e.Children[0].Name
	for i := 1; i < len(e.Children); i++ {
		if name != e.Children[i].Name {
			return false
		}
	}
	return true
}

func (e *xmlNode) GetFirstElement(name string) *xmlNode {
	for _, c := range e.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (e *xmlNode) writeStart(writer *bufio.Writer) {
	writer.WriteString("<")
	writer.WriteString(e.Name)
	for k, v := range e.Attributes {
		writer.WriteString(fmt.Sprintf(" %v=\"%v\"", k, v))
	}
	writer.WriteString(">")
}
