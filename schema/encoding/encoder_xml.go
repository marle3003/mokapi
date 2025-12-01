package encoding

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"net/url"
	"strings"
)

type XmlEncoder struct {
	w *bufio.Writer
}

func MarshalXml(data any, s *schema.Schema) ([]byte, error) {
	x := XmlEncoder{}
	return x.Encode(data, s)
}

func (x *XmlEncoder) Encode(data any, s *schema.Schema) ([]byte, error) {
	var buffer bytes.Buffer
	x.w = bufio.NewWriter(&buffer)

	rootName := "data"
	if s != nil && s.Id != "" {
		u, _ := url.Parse(s.Id)
		seg := strings.Split(u.Path, "/")
		if len(seg) > 1 {
			rootName = seg[len(seg)-1]
		}
	}

	_, err := x.w.WriteString(xml.Header)
	if err != nil {
		return nil, err
	}

	err = x.writeXmlElement(rootName, data)
	if err != nil {
		return nil, err
	}
	err = x.w.Flush()
	return buffer.Bytes(), err
}

func (x *XmlEncoder) writeXmlElement(name string, data any) error {
	_, err := x.w.WriteString(fmt.Sprintf("<%s>", name))
	if err != nil {
		return err
	}

	switch t := data.(type) {
	case []any:
		for _, item := range t {
			err = x.writeXmlElement("items", item)
			if err != nil {
				return err
			}
		}
	case map[string]any:
		for key, item := range t {
			err = x.writeXmlElement(key, item)
			if err != nil {
				return err
			}
		}
	case *sortedmap.LinkedHashMap[string, any]:
		for it := t.Iter(); it.Next(); {
			err = x.writeXmlElement(it.Key(), it.Value())
			if err != nil {
				return err
			}
		}
	default:
		_, err = x.w.WriteString(fmt.Sprintf("%v", data))
		if err != nil {
			return err
		}
	}

	_, err = x.w.WriteString(fmt.Sprintf("</%s>", name))
	return err
}
