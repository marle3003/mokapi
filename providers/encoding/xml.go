package encoding

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"mokapi/config/dynamic/openapi"
	"strconv"
	"strings"
)

func UnmarshalXml(s string, schema *openapi.SchemaRef) (interface{}, error) {
	data := []byte(s)
	e, err := decode(data)
	if err != nil {
		return nil, err
	}
	if len(e.Children) == 0 {
		return e, nil
	}
	if !isTypeDefined(schema) {
		return e.Children[0], nil
	}
	// skip root element
	obj, err := e.Children[0].parse(schema)
	return obj, err
}

func MarshalXML(v interface{}, schema *openapi.SchemaRef) ([]byte, error) {
	m := &StringMap{Data: v, Schema: schema}
	xmlString, err := xml.Marshal(m)
	if err != nil {
		return nil, err
	}
	return []byte(xml.Header + string(xmlString)), nil
}

type StringMap struct {
	Data   interface{}
	Schema *openapi.SchemaRef
}

func (m StringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	wrapped := false
	if m.Schema == nil || m.Schema.Value == nil {
		return errors.Errorf("unable to marshal xml with undefined schema")
	}

	switch m.Data.(type) {
	case []interface{}:
		if m.Schema.Value.Type != "array" {
			return fmt.Errorf("expected %q but found array", m.Schema.Value.Type)
		}
	case map[string]interface{}:
		if m.Schema.Value.Type != "object" {
			return fmt.Errorf("expected %q but found object", m.Schema.Value.Type)
		}
	}

	if m.Schema.Value.Xml != nil {
		wrapped = m.Schema.Value.Xml.Wrapped
		if m.Schema.Value.Xml.Name != "" {
			start.Name.Local = m.Schema.Value.Xml.Name
		}
		if m.Schema.Value.Xml.Prefix != "" {
			start.Name.Local = fmt.Sprintf("%s:%s", m.Schema.Value.Xml.Prefix, start.Name.Local)
		}

		if m.Schema.Value.Xml.Namespace != "" {
			start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "xmlns:" + m.Schema.Value.Xml.Prefix}, Value: m.Schema.Value.Xml.Namespace})
		}
	}

	if wrapped {
		err := e.EncodeToken(xml.StartElement{Name: xml.Name{Local: start.Name.Local}})
		if err != nil {
			return err
		}
	}

	if m.Schema.Value.Type == "array" {
		if list, ok := m.Data.([]interface{}); ok {
			for _, item := range list {
				propertyMap := &StringMap{Data: item, Schema: m.Schema.Value.Items}
				t := xml.StartElement{Name: start.Name}
				err := e.EncodeElement(propertyMap, t)
				if err != nil {
					return err
				}
				err = e.EncodeToken(xml.EndElement{Name: t.Name})
				if err != nil {
					return err
				}
			}
		}
	} else if m.Schema.Value.Type == "object" {
		o := m.Data.(map[string]interface{})

		// Attributes
		for propertyName, propertySchema := range m.Schema.Value.Properties.Value {
			// if property is mapped to attribute
			if propertySchema.Value.Xml == nil || !propertySchema.Value.Xml.Attribute {
				continue
			}

			attributeName := propertyName
			if propertySchema.Value.Xml.Name != "" {
				attributeName = propertySchema.Value.Xml.Name
			}
			if propertySchema.Value.Xml.Prefix != "" {
				attributeName = fmt.Sprintf("%s:%s", propertySchema.Value.Xml.Prefix, attributeName)
			}

			if p, ok := o[propertyName]; ok {
				start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: attributeName}, Value: fmt.Sprint(p)})
			}
		}

		err := e.EncodeToken(start)
		if err != nil {
			return err
		}

		err = encodeObject(e, o, m.Schema)
		if err != nil {
			return err
		}

		err = e.EncodeToken(xml.EndElement{Name: start.Name})
		if err != nil {
			return err
		}
	} else {
		value := fmt.Sprint(m.Data)
		if m.Schema.Value.Xml != nil && m.Schema.Value.Xml.CData {
			err := e.EncodeElement(struct {
				S string `xml:",innerxml"`
			}{
				S: "<![CDATA[" + fmt.Sprint(m.Data) + "]]>",
			}, start)
			if err != nil {
				return err
			}
		} else {
			err := e.EncodeElement(value, start)
			if err != nil {
				return err
			}
		}

		err := e.EncodeToken(xml.EndElement{Name: start.Name})
		if err != nil {
			return err
		}
	}

	if wrapped {
		err := e.EncodeToken(xml.EndElement{Name: start.Name})
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	err := e.Flush()
	if err != nil {
		return err
	}

	return nil
}

func encodeObject(e *xml.Encoder, obj map[string]interface{}, schema *openapi.SchemaRef) error {
	for propertyName, propertySchema := range schema.Value.Properties.Value {
		if propertySchema.Value.Xml != nil && propertySchema.Value.Xml.Attribute {
			continue
		}

		if p, ok := obj[propertyName]; ok {
			propertyMap := &StringMap{Data: p, Schema: propertySchema}
			t := xml.StartElement{Name: xml.Name{Local: propertyName}}
			err := e.EncodeElement(propertyMap, t)
			if err != nil {
				return err
			}
			err = e.EncodeToken(xml.EndElement{Name: t.Name})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type XmlNode struct {
	Name       string
	Children   []*XmlNode
	Attributes map[string]string
	Content    string
}

func (e *XmlNode) GetFirstElement(name string) *XmlNode {
	for _, c := range e.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func decode(data []byte) (*XmlNode, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	e := &XmlNode{Children: make([]*XmlNode, 0), Attributes: make(map[string]string)}
	err := e.decode(decoder)
	return e, err
}

func (e *XmlNode) decode(decoder *xml.Decoder) error {
	for {
		tok, err := decoder.Token()
		if err != nil && err != io.EOF {
			return err
		} else if tok == nil {
			break
		}

		switch tokEle := tok.(type) {
		case xml.StartElement:
			child := &XmlNode{Children: make([]*XmlNode, 0), Attributes: make(map[string]string)}
			child.Name = tokEle.Name.Local
			for _, a := range tokEle.Attr {
				child.Attributes[a.Name.Local] = a.Value
			}
			err := child.decode(decoder)
			if err != nil {
				return err
			}

			e.Children = append(e.Children, child)
		case xml.EndElement:
			return nil
		case xml.CharData:
			s := string(tokEle)
			e.Content = strings.TrimSuffix(strings.TrimSpace(s), "\n")
			_ = s
		}
	}

	return nil
}

func (e *XmlNode) parse(s *openapi.SchemaRef) (interface{}, error) {
	if s == nil || s.Value == nil || s.Value.Type == "string" {
		if e == nil {
			return "", nil
		}
		return e.Content, nil
	}

	switch s.Value.Type {
	case "object":
		props := map[string]interface{}{}
		for name, property := range s.Value.Properties.Value {
			xmlName := name
			if property.Value.Xml != nil && len(property.Value.Xml.Name) > 0 {
				xmlName = property.Value.Xml.Name
			}
			if property.Value.Xml != nil && property.Value.Xml.Attribute {
				if v, ok := e.Attributes[xmlName]; ok {
					props[name] = v
				} else if isPropertyRequired(name, s) {
					return nil, errors.Errorf("required property with name '%v' not found", name)
				}
			} else {
				c := e.GetFirstElement(xmlName)
				if c == nil {
					if isPropertyRequired(name, s) {
						return nil, errors.Errorf("required property with name '%v' not found", name)
					} else {
						continue
					}
				}
				v, err := c.parse(property)
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
		if s.Value.Xml != nil && s.Value.Xml.Wrapped {
			if len(e.Children) == 0 {
				return make([]interface{}, 0), nil
			}
			elements = e.Children[0].Children
		}
		array := make([]interface{}, len(elements))
		for i, item := range e.Children {
			v, err := item.parse(s.Value.Items)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to parse array element of type '%v'", s.Value.Items.Value.Type)
			}
			array[i] = v
		}
		return array, nil
	case "number":
		f, err := strconv.ParseFloat(e.Content, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse number '%v'", e.Content)
		}
		return f, nil
	case "integer":
		i, err := strconv.Atoi(e.Content)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse integer '%v'", e.Content)
		}
		return i, nil
	case "boolean":
		b, err := strconv.ParseBool(e.Content)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse bool '%v'", e.Content)
		}
		return b, nil
	}

	return nil, nil
}

func isPropertyRequired(name string, s *openapi.SchemaRef) bool {
	if s.Value.Required == nil {
		return false
	}
	for _, p := range s.Value.Required {
		if p == name {
			return true
		}
	}
	return false
}

func isTypeDefined(s *openapi.SchemaRef) bool {
	return s != nil && s.Value != nil && len(s.Value.Type) > 0
}
