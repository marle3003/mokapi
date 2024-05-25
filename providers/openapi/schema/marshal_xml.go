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

func writeXmlElement(name string, i interface{}, r *Ref, w io.Writer) {
	var s *Schema
	if r != nil {
		s = r.Value
	}

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
		var sItems *Ref
		if s != nil {
			sItems = s.Items
		}
		for _, item := range v {
			if item == nil {
				continue
			}
			writeXmlElement(name, item, sItems, w)
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
		attrs.Merge(getAttributes(v, r))
		writeXmlStart(w, name, attrs)

		for it := v.Iter(); it.Next(); {
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

func getAttributes(m *sortedmap.LinkedHashMap[string, interface{}], r *Ref) *sortedmap.LinkedHashMap[string, string] {
	attrs := &sortedmap.LinkedHashMap[string, string]{}
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
