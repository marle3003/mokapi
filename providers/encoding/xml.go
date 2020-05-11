package encoding

import (
	"encoding/xml"
	"fmt"
	"mokapi/service"
)

func MarshalXML(v interface{}, schema *service.Schema) ([]byte, error) {
	m := &StringMap{Data: v, Schema: schema}
	xmlString, error := xml.Marshal(m)
	if error != nil {
		return nil, error
	}
	return []byte(xml.Header + string(xmlString)), nil
}

type StringMap struct {
	Data   interface{}
	Schema *service.Schema
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

func encodeObject(e *xml.Encoder, obj map[string]interface{}, schema *service.Schema) {
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
