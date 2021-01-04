package encoding

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"mokapi/models"
	"strconv"
)

func UnmarshalXml(s string, schema *models.Schema) (interface{}, error) {
	data := []byte(s)
	e, err := decode(data)
	if err != nil {
		return nil, err
	}
	if len(e.elements) == 0 {
		return nil, nil
	}
	// skip root element
	obj, err := e.elements[0].parse(schema)
	return obj, err
}

func MarshalXML(v interface{}, schema *models.Schema) ([]byte, error) {
	m := &StringMap{Data: v, Schema: schema}
	xmlString, error := xml.Marshal(m)
	if error != nil {
		return nil, error
	}
	return []byte(xml.Header + string(xmlString)), nil
}

type StringMap struct {
	Data   interface{}
	Schema *models.Schema
}

func (m StringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	wrapped := false
	if m.Schema.Xml != nil {
		wrapped = m.Schema.Xml.Wrapped
		if m.Schema.Xml.Name != "" {
			start.Name.Local = m.Schema.Xml.Name
		}
		if m.Schema.Xml.Prefix != "" {
			start.Name.Local = fmt.Sprintf("%s:%s", m.Schema.Xml.Prefix, start.Name.Local)
		}
		start.Name.Space = m.Schema.Xml.Namespace
	}

	if wrapped {
		e.EncodeToken(xml.StartElement{Name: xml.Name{Space: start.Name.Space, Local: start.Name.Local}})
	}

	if m.Schema.Type == "array" {
		if list, ok := m.Data.([]interface{}); ok {
			for _, item := range list {
				propertyMap := &StringMap{Data: item, Schema: m.Schema.Items}
				t := xml.StartElement{Name: start.Name}
				e.EncodeElement(propertyMap, t)
				e.EncodeToken(xml.EndElement{Name: t.Name})
			}
		}
	} else if m.Schema.Type == "object" {
		o := m.Data.(map[string]interface{})

		// Attributes
		for propertyName, propertySchema := range m.Schema.Properties {
			// if property is mapped to attribute
			if propertySchema.Xml == nil || !propertySchema.Xml.Attribute {
				continue
			}

			attributeName := propertyName
			if propertySchema.Xml.Name != "" {
				attributeName = propertySchema.Xml.Name
			}
			if propertySchema.Xml.Prefix != "" {
				attributeName = fmt.Sprintf("%s:%s", propertySchema.Xml.Prefix, attributeName)
			}

			if p, ok := o[propertyName]; ok {
				start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: propertySchema.Xml.Namespace, Local: attributeName}, Value: fmt.Sprint(p)})
			}
		}

		e.EncodeToken(start)

		encodeObject(e, o, m.Schema)

		e.EncodeToken(xml.EndElement{Name: start.Name})
	} else {
		value := fmt.Sprint(m.Data)
		if m.Schema.Xml != nil && m.Schema.Xml.CData {
			e.EncodeElement(struct {
				S string `xml:",innerxml"`
			}{
				S: "<![CDATA[" + string(fmt.Sprint(m.Data)) + "]]>",
			}, start)
		} else {
			e.EncodeElement(value, start)
		}

		e.EncodeToken(xml.EndElement{Name: start.Name})
	}

	if wrapped {
		e.EncodeToken(xml.EndElement{Name: start.Name})
	}

	// flush to ensure tokens are written
	err := e.Flush()
	if err != nil {
		return err
	}

	return nil
}

func encodeObject(e *xml.Encoder, obj map[string]interface{}, schema *models.Schema) {
	for propertyName, propertySchema := range schema.Properties {
		if propertySchema.Xml != nil && propertySchema.Xml.Attribute {
			continue
		}

		if p, ok := obj[propertyName]; ok {
			propertyMap := &StringMap{Data: p, Schema: propertySchema}
			t := xml.StartElement{Name: xml.Name{Local: propertyName}}
			e.EncodeElement(propertyMap, t)
			e.EncodeToken(xml.EndElement{Name: t.Name})
		}
	}
}

type element struct {
	name       string
	elements   []*element
	attributes map[string]string
	content    string
}

func (e *element) GetFirstElement(name string) *element {
	for _, c := range e.elements {
		if c.name == name {
			return c
		}
	}
	return nil
}

func decode(data []byte) (*element, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	e := &element{elements: make([]*element, 0), attributes: make(map[string]string)}
	err := e.decode(decoder)
	return e, err
}

func (e *element) decode(decoder *xml.Decoder) error {
	for {
		tok, err := decoder.Token()
		if err != nil && err != io.EOF {
			return err
		} else if tok == nil {
			break
		}

		switch tokEle := tok.(type) {
		case xml.StartElement:
			child := &element{elements: make([]*element, 0), attributes: make(map[string]string)}
			child.name = tokEle.Name.Local
			for _, a := range tokEle.Attr {
				child.attributes[a.Name.Local] = a.Value
			}
			child.decode(decoder)
			e.elements = append(e.elements, child)
		case xml.EndElement:
			return nil
		case xml.CharData:
			s := string(tokEle)
			e.content = s
			_ = s
		}
	}

	return nil
}

func (e *element) parse(s *models.Schema) (interface{}, error) {
	if s == nil || s.Type == "string" {
		return e.content, nil
	}

	switch s.Type {
	case "object":
		props := map[string]interface{}{}
		for name, property := range s.Properties {
			if property.Xml != nil && len(property.Xml.Name) > 0 {
				name = property.Xml.Name
			}
			if property.Xml != nil && property.Xml.Attribute {
				if v, ok := e.attributes[name]; ok {
					props[name] = v
				} else if s.IsPropertyRequired(name) {
					return nil, errors.Errorf("required property with name '%v' not found", name)
				}
			} else {
				c := e.GetFirstElement(name)
				if c == nil && s.IsPropertyRequired(name) {
					return nil, errors.Errorf("required property with name '%v' not found", name)
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
		elements := e.elements
		if s.Xml != nil && s.Xml.Wrapped {
			if len(e.elements) == 0 {
				return make([]interface{}, 0), nil
			}
			elements = e.elements[0].elements
		}
		array := make([]interface{}, len(elements))
		for i, item := range e.elements {
			v, err := item.parse(s.Items)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to parse array element of type '%v'", s.Items.Type)
			}
			array[i] = v
		}
		return array, nil
	case "number":
		f, err := strconv.ParseFloat(e.content, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse number '%v'", e.content)
		}
		return f, nil
	case "integer":
		i, err := strconv.Atoi(e.content)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse integer '%v'", e.content)
		}
		return i, nil
	case "boolean":
		b, err := strconv.ParseBool(e.content)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse bool '%v'", e.content)
		}
		return b, nil
	}

	return nil, nil
}
