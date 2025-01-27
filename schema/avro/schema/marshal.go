package schema

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"slices"
)

func (s *Schema) Marshal(v interface{}) ([]byte, error) {
	p := Parser{Schema: s}
	_, err := p.Parse(v)
	if err != nil {
		return nil, err
	}
	w := bytes.Buffer{}
	b := make([]byte, 32)

	writeUnionIndex := func(typeName string) {
		if len(s.Type) <= 1 {
			return
		}
		if index, ok := s.Is(typeName); ok {
			n := binary.PutVarint(b, int64(index))
			w.Write(b[:n])
		}
	}

	switch t := v.(type) {
	case nil:
		writeUnionIndex("null")
	case bool:
		writeUnionIndex("boolean")
		if t {
			w.WriteByte(1)
		} else {
			w.WriteByte(0)
		}
	case int:
		writeUnionIndex("int")
		n := binary.PutVarint(b, int64(t))
		w.Write(b[:n])
	case int64:
		writeUnionIndex("long")
		n := binary.PutVarint(b, t)
		w.Write(b[:n])
	case float32:
		writeUnionIndex("float")
		binary.LittleEndian.PutUint32(b, math.Float32bits(t))
		w.Write(b[:4])
	case float64:
		writeUnionIndex("double")
		binary.LittleEndian.PutUint64(b, math.Float64bits(t))
		w.Write(b[:8])
	case string:
		if _, ok := s.Is("enum"); ok {
			writeUnionIndex("enum")
			index := slices.Index(s.Symbols, t)
			if index == -1 {
				return nil, fmt.Errorf(`invalid enum type "%s"`, t)
			}
			n := binary.PutVarint(b, int64(index))
			w.Write(b[:n])
		} else {
			writeUnionIndex("string")
			n := binary.PutVarint(b, int64(len(t)))
			w.Write(b[:n])
			w.WriteString(t)
		}
	case map[string]interface{}:
		if _, ok := s.Is("map"); ok {
			writeUnionIndex("map")
			n := binary.PutVarint(b, int64(len(t)))
			w.Write(b[:n])
			for key, val := range t {
				n = binary.PutVarint(b, int64(len(key)))
				w.Write(b[:n])
				w.WriteString(key)
				bVal, err := s.Values.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf(`marshal value of "%s" failed: %v`, key, err)
				}
				w.Write(bVal)
			}
		} else {
			writeUnionIndex("record")
			for _, field := range s.Fields {
				fv, ok := t[field.Name]
				if !ok {
					return nil, fmt.Errorf(`field "%s" not found`, field.Name)
				}
				fb, err := field.Marshal(fv)
				if err != nil {
					return nil, fmt.Errorf(`marshal field "%s" failed: %w`, field.Name, err)
				}
				w.Write(fb)
			}
		}
	case []interface{}:
		writeUnionIndex("array")
		n := binary.PutVarint(b, int64(len(t)))
		w.Write(b[:n])
		for _, i := range t {
			bItem, err := s.Items.Marshal(i)
			if err != nil {
				return nil, err
			}
			w.Write(bItem)
		}
	}

	if w.Len() == 0 {
		return []byte{}, nil
	}
	return w.Bytes(), nil
}

func (s *Schema) Is(typeName string) (int, bool) {
	for index, t := range s.Type {
		switch v := t.(type) {
		case string:
			if v == typeName {
				return index, true
			}
		case *Schema:
			if _, ok := v.Is(typeName); ok {
				return index, true
			}
		}
	}
	return -1, false
}
