package encoding

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"mokapi/config/dynamic/openapi"
	"net/url"
	"strings"
)

type xmlEncoder struct {
	writer *bufio.Writer
}

type startElement struct {
	name  string
	space string
	attr  map[string]string
}

func newXmlWriter(w io.Writer) *xmlEncoder {
	return &xmlEncoder{writer: bufio.NewWriter(w)}
}

func (w *xmlEncoder) write(i interface{}, schema *openapi.SchemaRef) error {
	w.writer.WriteString("<?xml version=\"1.0\"?>")
	if len(schema.Ref) > 0 {
		u, _ := url.Parse(schema.Ref)
		seg := strings.Split(u.Fragment, "/")
		return w.encodeElement(seg[len(seg)-1], i, schema)
	}

	return w.encodeElement("", i, schema)
}

func (w *xmlEncoder) encodeElement(name string, i interface{}, schema *openapi.SchemaRef) error {
	if schema == nil || schema.Value == nil {
		return errors.Errorf("unable to marshal xml with undefined schema")
	}

	start := startElement{
		name: name,
		attr: make(map[string]string),
	}
	wrapped := false
	if schema.Value.Xml != nil {
		wrapped = schema.Value.Xml.Wrapped
		if len(schema.Value.Xml.Name) > 0 {
			start.name = schema.Value.Xml.Name
		}
		if schema.Value.Xml.Prefix != "" {
			start.name = fmt.Sprintf("%s:%s", schema.Value.Xml.Prefix, start.name)
		}
		if schema.Value.Xml.Namespace != "" {
			start.attr["xmlns:"+schema.Value.Xml.Prefix] = schema.Value.Xml.Namespace
		}
	}

	if len(start.name) == 0 {
		return fmt.Errorf("missing xml element name")
	}

	if wrapped {
		_, err := w.writer.WriteString(fmt.Sprintf("<%v>", start.name))
		if err != nil {
			return err
		}
	}

	switch schema.Value.Type {
	case "array":
		if list, ok := i.([]interface{}); ok {
			for _, item := range list {
				err := w.encodeElement(start.name, item, schema.Value.Items)
				if err != nil {
					return err
				}
			}
		} else if list, ok := i.(map[interface{}]interface{}); ok {
			for _, item := range list {
				err := w.encodeElement(start.name, item, schema.Value.Items)
				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("expected array but found %T", i)
		}
	case "object":
		o, ok := i.(map[string]interface{})
		if !ok {
			if m, ok := i.(map[interface{}]interface{}); ok {
				o = make(map[string]interface{})
				for k, v := range m {
					o[fmt.Sprintf("%v", k)] = v
				}
			} else {
				return fmt.Errorf("expected object but found %T", i)
			}
		}
		addAttribute(start, o, schema.Value)

		w.writeStart(start)

		for propertyName, propertySchema := range schema.Value.Properties.Value {
			if propertySchema.Value.Xml != nil && propertySchema.Value.Xml.Attribute {
				continue
			}
			err := w.encodeElement(propertyName, o[propertyName], propertySchema)
			if err != nil {
				return err
			}
		}

		w.writer.WriteString(fmt.Sprintf("</%v>", start.name))
	default:
		w.writeStart(start)

		value := toString(i)
		if schema.Value.Xml != nil && schema.Value.Xml.CData {
			w.writer.WriteString(fmt.Sprintf("<![CDATA[%v]]>", value))
		} else {
			w.writer.WriteString(value)
		}

		w.writer.WriteString(fmt.Sprintf("</%v>", start.name))
	}

	if wrapped {
		w.writer.WriteString(fmt.Sprintf("</%v>", start.name))
	}

	return w.writer.Flush()
}

func (w *xmlEncoder) writeStart(start startElement) {
	w.writer.WriteString("<")
	w.writer.WriteString(start.name)
	for k, v := range start.attr {
		w.writer.WriteString(fmt.Sprintf(" %v=\"%v\"", k, toString(v)))
	}
	w.writer.WriteString(">")
}

func addAttribute(start startElement, obj map[string]interface{}, schema *openapi.Schema) {
	for propertyName, propertySchema := range schema.Properties.Value {
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

		if p, ok := obj[propertyName]; ok {
			start.attr[attributeName] = toString(p)
		}
	}
}
