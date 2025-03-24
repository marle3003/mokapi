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

	w := NewWriter()
	err = w.Write(v, s)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

type Writer struct {
	w bytes.Buffer
	b []byte
}

func NewWriter() *Writer {
	return &Writer{b: make([]byte, 32)}
}

func (w *Writer) Write(v interface{}, s *Schema) error {
	var typeIndexes []int
	var ok bool
	var ms *Schema

	if typeIndexes, ms, ok = getMatchingSchema(v, s); ok {
		s = ms
	} else {
		return fmt.Errorf("value '%v' does not match schema %v", v, s)
	}

	if len(typeIndexes) > 0 {
		for i, val := range typeIndexes {
			if i == len(typeIndexes)-1 && len(s.Type) == 1 {
				// write last index only if there are more types
				break
			}
			n := binary.PutVarint(w.b, int64(val))
			w.w.Write(w.b[:n])
		}
	}
	typeIndex := typeIndexes[len(typeIndexes)-1]

	switch s.Type[typeIndex] {
	case "boolean":
		b := v.(bool)
		if b {
			w.w.WriteByte(1)
		} else {
			w.w.WriteByte(0)
		}
	case "int":
		i, err := toInt64(v)
		if err != nil {
			return fmt.Errorf("value %v does not match schema %v", v, s)
		}
		n := binary.PutVarint(w.b, i)
		if n > 5 {
			return fmt.Errorf("value %v does not match schema %v", v, s)
		}
		w.w.Write(w.b[:n])
	case "long":
		i, err := toInt64(v)
		if err != nil {
			return fmt.Errorf("value %v does not match schema %v", v, s)
		}
		n := binary.PutVarint(w.b, i)
		w.w.Write(w.b[:n])
	case "float":
		f := v.(float32)
		binary.LittleEndian.PutUint32(w.b, math.Float32bits(f))
		w.w.Write(w.b[:4])
	case "double":
		d := v.(float64)
		binary.LittleEndian.PutUint64(w.b, math.Float64bits(d))
		w.w.Write(w.b[:8])
	case "string":
		str := v.(string)
		n := binary.PutVarint(w.b, int64(len(str)))
		w.w.Write(w.b[:n])
		w.w.WriteString(str)
	case "enum":
		str := v.(string)
		index := slices.Index(s.Symbols, str)
		if index == -1 {
			return fmt.Errorf("value '%v' does not match one in the symbols %v", v, ToString(s.Symbols))
		}
		n := binary.PutVarint(w.b, int64(index))
		w.w.Write(w.b[:n])
	case "record":
		m := v.(map[string]interface{})
		for _, field := range s.Fields {
			fv, ok := m[field.Name]
			if !ok {
				return fmt.Errorf(`field "%s" not found`, field.Name)
			}
			err := w.Write(fv, field)
			if err != nil {
				return fmt.Errorf(`marshal field "%s" failed: %w`, field.Name, err)
			}
		}
	case "map":
		m := v.(map[string]interface{})
		n := binary.PutVarint(w.b, int64(len(m)))
		w.w.Write(w.b[:n])
		for key, val := range m {
			n = binary.PutVarint(w.b, int64(len(key)))
			w.w.Write(w.b[:n])
			w.w.WriteString(key)
			err := w.Write(val, s.Values)
			if err != nil {
				return fmt.Errorf(`marshal value of "%s" failed: %v`, key, err)
			}
		}
	case "array":
		a := v.([]interface{})
		n := binary.PutVarint(w.b, int64(len(a)))
		w.w.Write(w.b[:n])
		for _, i := range a {
			err := w.Write(i, s.Items)
			if err != nil {
				return err
			}
		}
		w.w.WriteByte(0)
	case "fixed":
		if str, ok := v.(string); ok {
			w.w.WriteString(str)
		} else {
			w.w.Write(v.([]byte))
		}
	}

	return nil
}

func (w *Writer) Bytes() []byte {
	if w.w.Len() == 0 {
		return []byte{}
	}
	return w.w.Bytes()
}

func getMatchingSchema(v interface{}, s *Schema) ([]int, *Schema, bool) {
	for index, t := range s.Type {
		switch vt := t.(type) {
		case string:
			if match(v, vt) {
				switch vt {
				case "enum":
					if !slices.Contains(s.Symbols, v.(string)) {
						continue
					}
				case "fixed":
					if str, ok := v.(string); ok && len(str) != s.Size {
						continue
					}
					if b, ok := v.([]byte); ok && len(b) != s.Size {
						continue
					}
				}
				return []int{index}, s, true
			}
		case *Schema:
			if i2, s2, ok := getMatchingSchema(v, vt); ok {
				if len(s.Type) > 1 {
					i := []int{index}
					i = append(i, i2...)
					return i, s2, true
				}
				return i2, s2, true
			}
		}
	}
	return nil, nil, false
}

func match(v interface{}, typeName string) bool {
	switch typeName {
	case "null":
		return v == nil
	case "string", "enum":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "int", "long":
		switch v.(type) {
		case int, int8, int16, int32, int64:
			return true
		}
	case "float":
		_, ok := v.(float32)
		return ok
	case "double":
		_, ok := v.(float64)
		return ok
	case "record", "map":
		_, ok := v.(map[string]interface{})
		return ok
	case "array":
		_, ok := v.([]interface{})
		return ok
	case "fixed":
		if _, ok := v.(string); ok {
			return true
		}
		if _, ok := v.([]byte); ok {
			return true
		}
	}

	return false
}

func toInt64(v interface{}) (int64, error) {
	switch vi := v.(type) {
	case int:
		return int64(vi), nil
	case int8:
		return int64(vi), nil
	case int16:
		return int64(vi), nil
	case int32:
		return int64(vi), nil
	case int64:
		return vi, nil
	}
	return 0, fmt.Errorf("cannot convert %v to int", v)
}
