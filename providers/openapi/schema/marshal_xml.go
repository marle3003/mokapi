package schema

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"mokapi/sortedmap"
	"net/url"
	"strings"
)

func marshalXml(i interface{}, r *Schema) ([]byte, error) {
	if r == nil || r.SubSchema == nil {
		return nil, fmt.Errorf("no schema provided")
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	var name string
	if r.Xml != nil && len(r.Xml.Name) > 0 {
		name = r.Xml.Name
	} else if len(r.Ref) > 0 {
		u, _ := url.Parse(r.Ref)
		seg := strings.Split(u.Fragment, "/")
		name = seg[len(seg)-1]
	} else {
		return nil, fmt.Errorf("root element name is undefined: reference name of schema and attribute xml.name is empty")
	}

	if i == nil {
		_, _ = writer.WriteString(fmt.Sprintf("<%v></%v>", name, name))
	} else {
		writeXmlElement(name, i, r, writer)
	}

	err := writer.Flush()
	return buffer.Bytes(), err
}

func writeXmlElement(name string, i interface{}, s *Schema, w io.Writer) {
	wrapped := false
	attrs := &sortedmap.LinkedHashMap[string, string]{}
	if s != nil && s.Xml != nil {
		wrapped = s.Xml.Wrapped
		if len(s.Xml.Name) > 0 {
			name = s.Xml.Name
		}
		if s.Xml.Prefix != "" {
			name = fmt.Sprintf("%s:%s", s.Xml.Prefix, name)
		}
		if s.Xml.Namespace != "" {
			attrName := "xmlns"
			if len(s.Xml.Prefix) > 0 {
				attrName += fmt.Sprintf(":%v", s.Xml.Prefix)
			}
			attrs.Set(attrName, s.Xml.Namespace)
		}
	}

	switch v := i.(type) {
	case []interface{}:
		if wrapped {
			writeXmlStart(w, name, attrs)
		}
		var items *Schema
		if s != nil {
			items = s.Items
		}
		for _, item := range v {
			if item == nil {
				continue
			}
			writeXmlElement(name, item, items, w)
		}
		if wrapped {
			writeXmlEnd(w, name)
		}
	case map[string]interface{}:
		writeXmlStart(w, name, attrs)
		for name, v := range v {
			writeXmlElement(name, v, nil, w)
		}
		writeXmlEnd(w, name)
	case *sortedmap.LinkedHashMap[string, interface{}]:
		attrs.Merge(getAttributes(v, s))
		writeXmlStart(w, name, attrs)

		for it := v.Iter(); it.Next(); {
			if it.Value() == nil {
				continue
			}

			propName := it.Key()
			prop := s.Properties.Get(propName)
			if prop != nil {
				x := prop.Xml
				if x != nil && x.Attribute {
					continue
				}
			}

			writeXmlElement(propName, it.Value(), prop, w)
		}

		writeXmlEnd(w, name)
	default:
		if i == nil {
			return
		}
		writeXmlStart(w, name, attrs)
		writeXmlContent(w, fmt.Sprintf("%v", i))
		writeXmlEnd(w, name)
	}
}

func getAttributes(m *sortedmap.LinkedHashMap[string, interface{}], r *Schema) *sortedmap.LinkedHashMap[string, string] {
	attrs := &sortedmap.LinkedHashMap[string, string]{}
	for it := m.Iter(); it.Next(); {
		if it.Value() == nil {
			continue
		}

		prop := r.Properties.Get(it.Key())
		if prop == nil || prop.Xml == nil || !prop.Xml.Attribute {
			continue
		}

		attrName := it.Key()
		if len(prop.Xml.Name) > 0 {
			attrName = prop.Xml.Name
		}
		if len(prop.Xml.Prefix) > 0 {
			attrName = fmt.Sprintf("%s:%s", prop.Xml.Prefix, attrName)
		}

		attrs.Set(attrName, fmt.Sprintf("%v", it.Value()))
	}
	return attrs
}

func writeXmlStart(w io.Writer, name string, attrs *sortedmap.LinkedHashMap[string, string]) {
	writeString(w, "<"+name)

	for it := attrs.Iter(); it.Next(); {
		writeString(w, fmt.Sprintf(" %v=\"%v\"", it.Key(), it.Value()))
	}

	writeString(w, ">")
}

func writeXmlContent(w io.Writer, s string) {
	_ = xml.EscapeText(w, []byte(s))
}

func writeXmlEnd(w io.Writer, name string) {
	writeString(w, fmt.Sprintf("</%v>", name))
}

func writeString(w io.Writer, s string) {
	_, _ = w.Write([]byte(s))
}
