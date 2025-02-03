package schema

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var NoSchemaId = fmt.Errorf("no schema id")

type Parser struct {
	Schema *Schema
}

func (p *Parser) Parse(data interface{}) (interface{}, error) {
	if b, ok := data.([]byte); ok {
		return p.parseFromByte(b)
	}
	return p.parseFromInterface(data)
}

func (p *Parser) parseFromByte(b []byte) (interface{}, error) {
	r := bytes.NewReader(b)

	_, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	// Check magic byte, if set, read version
	magic, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if magic == 0 {
		var version int32
		err = binary.Read(r, binary.BigEndian, &version)
		if err != nil {
			return nil, err
		}
	} else {
		_ = r.UnreadByte()
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
		if int(n) >= len(s.Type) {
			return nil, fmt.Errorf("index %v out of range in union at offset %v", r.Size()-int64(r.Len()), n)
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
		if n < 0 {
			return nil, fmt.Errorf("invalid byte length at offset %v: %d", r.Size()-int64(r.Len()), n)
		}
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
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			// todo: If a block’s count is negative, its absolute value is used, and the count is followed immediately
			// by a long block size indicating the number of bytes in the block.
			return nil, fmt.Errorf("invalid array length at offset %v: %v", r.Size()-int64(r.Len()), n)
		}
		a := make([]interface{}, 0, n)
		if n == 0 {
			return a, nil
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
			return nil, fmt.Errorf("invalid array end at offset %v", r.Size()-int64(r.Len()))
		}
		return a, nil
	case "enum":
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		if n < 0 || int(n) > len(s.Symbols) {
			return nil, fmt.Errorf("index %v out of enum range at offset %v", n, r.Size()-int64(r.Len()))
		}
		return s.Symbols[n], nil
	case "map":
		m := make(map[string]interface{})
		n, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid map length at offset %v: %v", r.Size()-int64(r.Len()), n)
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
		if s.Size < 0 {
			return nil, fmt.Errorf("invalid fixed size at offset %v: %v", r.Size()-int64(r.Len()), s.Size)
		}
		b := make([]byte, s.Size)
		_, err := r.Read(b)
		return b, err
	}

	return nil, fmt.Errorf("unknown schema type at offset %v: %s", r.Size()-int64(r.Len()), s.Type)
}

func readString(r *bytes.Reader) (string, error) {
	n, err := binary.ReadVarint(r)
	if err != nil {
		return "", err
	}
	if n < 0 {
		return "", fmt.Errorf("invalid string length at offset %v: %d", r.Size()-int64(r.Len()), n)
	}
	b := make([]byte, n)
	_, err = r.Read(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
