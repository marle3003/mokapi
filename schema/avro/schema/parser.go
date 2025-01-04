package schema

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Parser struct {
	Schema *Schema
}

func (p *Parser) Parse(data interface{}) (interface{}, error) {
	if b, ok := data.([]byte); ok {
		return p.parseFromByte(b)
	}
	return data, nil
}

func (p *Parser) parseFromByte(b []byte) (interface{}, error) {
	r := bytes.NewReader(b)

	magic, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	// Check magic byte, if set, read version
	if magic == 0 {
		// read magic byte
		_, _ = r.ReadByte()

		var version int32
		err = binary.Read(r, binary.BigEndian, &version)
		if err != nil {
			return nil, err
		}
	}

	return p.parse(r, p.Schema)
}

func (p *Parser) parse(r *bytes.Reader, s *Schema) (interface{}, error) {
	t := s.Type[0]
	if len(s.Type) > 1 {
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		t = s.Type[n]
	}

	if wrapped, ok := t.(*Schema); ok {
		return p.parse(r, wrapped)
	}

	switch t {
	case "null":
		return nil, nil
	case "boolean":
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		return b != 0, nil
	case "int", "long":
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		return n, nil
	case "float":
		var bits float32
		err := binary.Read(r, binary.LittleEndian, &bits)
		if err != nil {
			return nil, err
		}
		return bits, nil
	case "double":
		var bits float64
		err := binary.Read(r, binary.LittleEndian, &bits)
		if err != nil {
			return nil, err
		}
		return bits, nil
	case "string":
		return readString(r)
	case "byte":
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		_ = n
		b := make([]byte, n)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return b, nil
	case "record":
		m := make(map[string]interface{})
		for _, f := range s.Fields {
			v, err := p.parse(r, &f)
			if err != nil {
				return nil, err
			}
			m[f.Name] = v
		}
		return m, nil
	case "array":
		var a []interface{}
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(n); i++ {
			var item interface{}
			item, err = p.parse(r, s.Items)
			if err != nil {
				return nil, err
			}
			a = append(a, item)
		}
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if b != byte(0) {
			return nil, fmt.Errorf("invalid array end")
		}
		return a, nil
	case "enum":
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		if n < 0 || int(n) > len(s.Symbols) {
			return nil, fmt.Errorf("index %v out of enum range", n)
		}
		return s.Symbols[n], nil
	case "map":
		m := make(map[string]interface{})
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(n); i++ {
			key, err := readString(r)
			if err != nil {
				return nil, err
			}
			val, err := p.parse(r, s.Values)
			if err != nil {
				return nil, err
			}
			m[key] = val
		}
		return m, nil
	case "fixed":
		b := make([]byte, s.Size)
		_, err := r.Read(b)
		return b, err
	}

	return nil, fmt.Errorf("unknown schema type: %s", s.Type)
}

func readString(r *bytes.Reader) (string, error) {
	n, err := binary.ReadVarint(r)
	if err != nil {
		return "", err
	}
	b := make([]byte, n)
	_, err = r.Read(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
