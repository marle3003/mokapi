package decoding

import (
	"mokapi/media"
	"mokapi/xml"
)

type XmlDecoder struct {
}

func (d *XmlDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.Subtype == "xml"
}

func (d *XmlDecoder) Decode(b []byte, _ media.ContentType, _ DecodeFunc) (i interface{}, err error) {
	n, err := xml.Read(b)
	if err != nil {
		return nil, err
	}
	if len(n.Children) == 0 {
		return nil, nil
	}
	return parseXml(n)
}

func parseXml(n *xml.Node) (interface{}, error) {
	if len(n.Content) > 0 {
		return n.Content, nil
	}

	if isArray(n) {
		result := make([]interface{}, 0, len(n.Children))
		for _, c := range n.Children {
			v, err := parseXml(c)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
		return result, nil
	} else {
		result := map[string]interface{}{}
		for it := n.Attributes.Iter(); it.Next(); {
			result[it.Key()] = it.Value()
		}
		for _, c := range n.Children {
			v, err := parseXml(c)
			if err != nil {
				return nil, err
			}
			result[c.Name] = v
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
